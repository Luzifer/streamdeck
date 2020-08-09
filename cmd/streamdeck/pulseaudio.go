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

func (p pulseAudioClient) GetSinkInputVolume(match string) (float64, error) {
	m, err := regexp.Compile(match)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to compile given match RegEx")
	}

	var resp proto.GetSinkInputInfoListReply
	if err := p.client.RawRequest(&proto.GetSinkInputInfoList{}, &resp); err != nil {
		return 0, errors.Wrap(err, "Unable to list sink inputs")
	}

	for _, info := range resp {
		if !m.MatchString(info.MediaName) && !m.Match(info.Properties["application.name"]) {
			continue
		}

		sinkBase, err := p.getSinkBaseVolumeByIndex(info.SinkIndex)
		if err != nil {
			return 0, errors.Wrap(err, "Unable to get sink base volume")
		}

		return p.unifyChannelVolumes(info.ChannelVolumes) / sinkBase, nil
	}

	return 0, errors.New("No such sink")
}

func (p pulseAudioClient) GetSinkVolume(match string) (float64, error) {
	m, err := regexp.Compile(match)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to compile given match RegEx")
	}

	var resp proto.GetSinkInfoListReply
	if err := p.client.RawRequest(&proto.GetSinkInfoList{}, &resp); err != nil {
		return 0, errors.Wrap(err, "Unable to list sinks")
	}

	for _, info := range resp {
		if !m.MatchString(info.SinkName) && !m.MatchString(info.Device) {
			continue
		}

		return p.unifyChannelVolumes(info.ChannelVolumes) / float64(info.BaseVolume), nil
	}

	return 0, errors.New("No such sink")
}

func (p pulseAudioClient) getSinkBaseVolumeByIndex(idx uint32) (float64, error) {
	var resp proto.GetSinkInfoReply
	if err := p.client.RawRequest(&proto.GetSinkInfo{SinkIndex: idx}, &resp); err != nil {
		return 0, errors.Wrap(err, "Unable to get sink")
	}

	return float64(resp.BaseVolume), nil
}

func (p pulseAudioClient) unifyChannelVolumes(v proto.ChannelVolumes) float64 {
	if len(v) == 0 {
		return 0
	}

	if len(v) == 1 {
		return float64(v[0])
	}

	var vMin = float64(v[0])
	for i := 1; i < len(v); i++ {
		vMin = math.Min(vMin, float64(v[i]))
	}

	return vMin
}
