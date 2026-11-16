package app

import (
	"context"
	"errors"
	"time"

	"github.com/egoist/mygo"

	"nextopsx/internal/i18n"
)

// AboutState is the settings page's "About" section: the running version and
// the state of an automatic-update check/install.
type AboutState struct {
	Version string
	Enabled bool

	Checking  bool
	Checked   bool
	Available bool
	Latest    string
	Notes     string
	Err       string

	Installing bool
	Downloaded int64
	Total      int64
	Installed  bool

	update *mygo.Update
}

// LoadAbout refreshes the version and whether this build can update itself.
func (a *App) LoadAbout() {
	v := mygo.App.Version()
	if v == "" {
		v = i18n.T("settings.about.devVersion")
	}
	a.Settings.About.Version = v
	a.Settings.About.Enabled = mygo.Updater.Enabled()
}

// CheckForUpdates asks the configured feed for a newer version. It runs off
// the UI thread and reports through AboutState.
func (a *App) CheckForUpdates() {
	about := &a.Settings.About
	if about.Checking {
		return
	}
	if !mygo.Updater.Enabled() {
		about.Checking = false
		about.Checked = true
		about.Err = i18n.T("settings.about.disabled")
		return
	}
	about.Checking = true
	about.Checked = false
	about.Err = ""
	about.Available = false
	about.update = nil

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		up, err := mygo.Updater.Check(ctx)
		a.Post(func() {
			about.Checking = false
			about.Checked = true
			if err != nil {
				if errors.Is(err, mygo.ErrUpdatesDisabled) {
					about.Enabled = false
					about.Err = i18n.T("settings.about.disabled")
				} else {
					about.Err = i18n.Tf("settings.about.checkFailed", err.Error())
				}
				return
			}
			if up == nil {
				return
			}
			about.Available = true
			about.Latest = up.Version
			about.Notes = up.Notes
			about.update = up
		})
	}()
}

// InstallUpdate downloads and installs the update found by CheckForUpdates.
// The new version runs after RelaunchApp (or the next launch).
func (a *App) InstallUpdate() {
	about := &a.Settings.About
	if about.update == nil || about.Installing || about.Installed {
		return
	}
	about.Installing = true
	about.Err = ""
	about.Downloaded = 0
	about.Total = 0
	up := about.update

	go func() {
		err := up.Install(context.Background(), func(downloaded, total int64) {
			a.Post(func() {
				about.Downloaded = downloaded
				about.Total = total
			})
		})
		a.Post(func() {
			about.Installing = false
			if err != nil {
				about.Err = i18n.Tf("settings.about.installFailed", err.Error())
				return
			}
			about.Installed = true
			about.Available = false
		})
	}()
}

// RelaunchApp starts the new version after an update has been installed.
func (a *App) RelaunchApp() {
	mygo.App.Relaunch()
}

// UpdatePercent is the download progress as a whole percentage, or -1 when
// the total size is not known yet.
func (a *App) UpdatePercent() int {
	about := a.Settings.About
	if about.Total <= 0 {
		return -1
	}
	p := int(about.Downloaded * 100 / about.Total)
	if p > 100 {
		p = 100
	}
	return p
}
