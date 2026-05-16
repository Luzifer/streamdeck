# 2.0.0 / 2026-05-16

* Breaking Changes
  * feat!: drop support for PulseAudio
  * chore!: drop support for deprecated `keys` array in `key_press` action

* New Features
  * feat: add http display and action
  * feat: add support for env-var references in config
  * feat: introduce timeouts for display commands

* Improvements
  * fix: resolve recurrent errors from exec displays

* Bugfixes
  * chore: replace deprecated YAML library
  * fix(deps): update github.com/sashko/go-uinput digest to faf003f
  * fix(deps): update module github.com/fsnotify/fsnotify to v1.10.1 (#5)
  * fix(deps): update module github.com/luzifer/rconfig/v2 to v2.6.2 (#6)
  * fix(deps): update module github.com/sirupsen/logrus to v1.9.4 (#4)
  * fix(deps): update module github.com/sstallion/go-hid to v0.15.0 (#7)
  * fix(deps): update module golang.org/x/image to v0.40.0 (#9)
  * refactor: split code into modules; split config per module

# 1.7.2 / 2024-10-29

  * Update Go dependencies

# 1.7.1 / 2023-10-15

  * Update dependencies

# 1.7.0 / 2022-10-06

  * Add support for StreamDeck Mini V2
  * Add `background_color` attribute to `exec` display elements
  * Add `background_color` attribute to `text` display elements

# 1.6.0 / 2022-02-05

  * Add support for StreamDeck Mini (#12) (Thanks to @mcrute)

# 1.5.0 / 2021-05-27

  * Add caption support for image buttons

# 1.4.0 / 2021-05-05

  * Add Meta/Windows/Super\_L modifier (#10) (Thanks @pheerai)
  * Prevent system crash by too fast executions
  * Move configuration to more stable format (#8)

# 1.3.0 / 2021-02-26

  * Add "text" display element

# 1.2.2 / 2021-01-01

  * [#4] Use strict config parsing

# 1.2.1 / 2020-11-17

  * Fix: Use proper context for delayed errors
  * Fix: Do not pile up the same page on refresh

# 1.2.0 / 2020-10-30

  * Fix blank page not doing anything
  * Implement relative movement through page-stack

# 1.1.0 / 2020-10-20

  * Add support for short / long press actions

# 1.0.0 / 2020-09-20

  * Initial release
