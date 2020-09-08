// +build linux

package main

import (
	"math"
	"regexp"

	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var pulseClient *pulseAudioClient

func init() {
	var err error
	if pulseClient, err = newPulseAudioClient(); err != nil {
		log.WithError(err).Error("Unable to connect to PulseAudio, functionality will not work")
	}
}

var errPulseNoSuchDevice = errors.New("No such device")

type pulseAudioClient struct {
	client *pulse.Client
}

func newPulseAudioClient() (*pulseAudioClient, error) {
	c, err := pulse.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create pulse client")
	}

	return &pulseAudioClient{client: c}, nil
}

func (p pulseAudioClient) Close() { p.client.Close() }

func (p pulseAudioClient) GetSinkInputVolume(match string) (vol float64, muted bool, idx []uint32, max uint32, err error) {
	m, err := regexp.Compile(match)
	if err != nil {
		return 0, false, nil, 0, errors.Wrap(err, "Unable to compile given match RegEx")
	}

	var resp proto.GetSinkInputInfoListReply
	if err := p.client.RawRequest(&proto.GetSinkInputInfoList{}, &resp); err != nil {
		return 0, false, nil, 0, errors.Wrap(err, "Unable to list sink inputs")
	}

	for _, info := range resp {
		if !m.MatchString(info.MediaName) && !m.Match(info.Properties["application.name"]) || info.Corked {
			continue
		}

		sinkBase, err := p.getSinkReferenceVolumeByIndex(info.SinkIndex)
		if err != nil {
			return 0, false, nil, 0, errors.Wrap(err, "Unable to get sink base volume")
		}

		if max != 0 && sinkBase != max {
			return 0, false, nil, 0, errors.New("found different sink bases")
		}

		idx = append(idx, info.SinkInputIndex)
		max = sinkBase
		muted = muted || info.Muted
		vol = math.Max(vol, p.unifyChannelVolumes(info.ChannelVolumes)/float64(sinkBase))
	}

	if len(idx) == 0 {
		return 0, false, nil, 0, errPulseNoSuchDevice
	}

	return vol, muted, idx, max, nil
}

func (p pulseAudioClient) GetSinkVolume(match string) (vol float64, muted bool, idx uint32, max uint32, err error) {
	m, err := regexp.Compile(match)
	if err != nil {
		return 0, false, 0, 0, errors.Wrap(err, "Unable to compile given match RegEx")
	}

	var resp proto.GetSinkInfoListReply
	if err := p.client.RawRequest(&proto.GetSinkInfoList{}, &resp); err != nil {
		return 0, false, 0, 0, errors.Wrap(err, "Unable to list sinks")
	}

	for _, info := range resp {
		if !m.MatchString(info.SinkName) && !m.MatchString(info.Device) {
			continue
		}

		return p.unifyChannelVolumes(info.ChannelVolumes) / float64(info.NumVolumeSteps), info.Mute, info.SinkIndex, info.NumVolumeSteps, nil
	}

	return 0, false, 0, 0, errPulseNoSuchDevice
}

func (p pulseAudioClient) GetSourceVolume(match string) (vol float64, muted bool, idx uint32, max uint32, err error) {
	m, err := regexp.Compile(match)
	if err != nil {
		return 0, false, 0, 0, errors.Wrap(err, "Unable to compile given match RegEx")
	}

	var resp proto.GetSourceInfoListReply
	if err := p.client.RawRequest(&proto.GetSourceInfoList{}, &resp); err != nil {
		return 0, false, 0, 0, errors.Wrap(err, "Unable to list sources")
	}

	for _, info := range resp {
		if !m.MatchString(info.SourceName) && !m.MatchString(info.Device) {
			continue
		}

		return p.unifyChannelVolumes(info.ChannelVolumes) / float64(info.NumVolumeSteps), info.Mute, info.SourceIndex, info.NumVolumeSteps, nil
	}

	return 0, false, 0, 0, errPulseNoSuchDevice
}

func (p pulseAudioClient) SetSinkInputVolume(match string, mute string, vol float64, absolute bool) error {
	stateVol, stateMute, stateIdxs, stateSteps, err := p.GetSinkInputVolume(match)
	if err != nil {
		return errors.Wrap(err, "Unable to get current state of sink input")
	}

	var cmds []proto.RequestArgs

	for _, stateIdx := range stateIdxs {
		switch mute {
		case "true":
			cmds = append(cmds, &proto.SetSinkInputMute{SinkInputIndex: stateIdx, Mute: true})
		case "false":
			cmds = append(cmds, &proto.SetSinkInputMute{SinkInputIndex: stateIdx, Mute: false})
		case "toggle":
			cmds = append(cmds, &proto.SetSinkInputMute{SinkInputIndex: stateIdx, Mute: !stateMute})
		}

		if absolute && vol >= 0 {
			cmds = append(cmds, &proto.SetSinkInputVolume{SinkInputIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(vol * float64(stateSteps))}})
		} else if vol != 0 {
			cmds = append(cmds, &proto.SetSinkInputVolume{SinkInputIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(math.Max(0, stateVol+vol) * float64(stateSteps))}})
		}
	}

	for _, cmd := range cmds {
		if err := p.client.RawRequest(cmd, nil); err != nil {
			return errors.Wrap(err, "Unable to execute command")
		}
	}

	return nil
}

func (p pulseAudioClient) SetSinkVolume(match string, mute string, vol float64, absolute bool) error {
	stateVol, stateMute, stateIdx, stateSteps, err := p.GetSinkVolume(match)
	if err != nil {
		return errors.Wrap(err, "Unable to get current state of sink")
	}

	var cmds []proto.RequestArgs

	switch mute {
	case "true":
		cmds = append(cmds, &proto.SetSinkMute{SinkIndex: stateIdx, Mute: true})
	case "false":
		cmds = append(cmds, &proto.SetSinkMute{SinkIndex: stateIdx, Mute: false})
	case "toggle":
		cmds = append(cmds, &proto.SetSinkMute{SinkIndex: stateIdx, Mute: !stateMute})
	}

	if absolute && vol >= 0 {
		cmds = append(cmds, &proto.SetSinkVolume{SinkIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(vol * float64(stateSteps))}})
	} else if vol != 0 {
		cmds = append(cmds, &proto.SetSinkVolume{SinkIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(math.Max(0, stateVol+vol) * float64(stateSteps))}})
	}

	for _, cmd := range cmds {
		if err := p.client.RawRequest(cmd, nil); err != nil {
			return errors.Wrap(err, "Unable to execute command")
		}
	}

	return nil
}

func (p pulseAudioClient) SetSourceVolume(match string, mute string, vol float64, absolute bool) error {
	stateVol, stateMute, stateIdx, stateSteps, err := p.GetSourceVolume(match)
	if err != nil {
		return errors.Wrap(err, "Unable to get current state of source")
	}

	var cmds []proto.RequestArgs

	switch mute {
	case "true":
		cmds = append(cmds, &proto.SetSourceMute{SourceIndex: stateIdx, Mute: true})
	case "false":
		cmds = append(cmds, &proto.SetSourceMute{SourceIndex: stateIdx, Mute: false})
	case "toggle":
		cmds = append(cmds, &proto.SetSourceMute{SourceIndex: stateIdx, Mute: !stateMute})
	}

	if absolute && vol >= 0 {
		cmds = append(cmds, &proto.SetSourceVolume{SourceIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(vol * float64(stateSteps))}})
	} else if vol != 0 {
		cmds = append(cmds, &proto.SetSourceVolume{SourceIndex: stateIdx, ChannelVolumes: proto.ChannelVolumes{uint32(math.Max(0, stateVol+vol) * float64(stateSteps))}})
	}

	for _, cmd := range cmds {
		if err := p.client.RawRequest(cmd, nil); err != nil {
			return errors.Wrap(err, "Unable to execute command")
		}
	}

	return nil
}

func (p pulseAudioClient) getSinkReferenceVolumeByIndex(idx uint32) (uint32, error) {
	var resp proto.GetSinkInfoReply
	if err := p.client.RawRequest(&proto.GetSinkInfo{SinkIndex: idx}, &resp); err != nil {
		return 0, errors.Wrap(err, "Unable to get sink")
	}

	return resp.NumVolumeSteps, nil
}

func (p pulseAudioClient) unifyChannelVolumes(v proto.ChannelVolumes) float64 {
	if len(v) == 0 {
		return 0
	}

	if len(v) == 1 {
		return float64(v[0])
	}

	vMin := float64(v[0])
	for i := 1; i < len(v); i++ {
		vMin = math.Min(vMin, float64(v[i]))
	}

	return vMin
}
