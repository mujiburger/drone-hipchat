package main

import (
	"fmt"
	"os"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
	"github.com/drone/drone-go/template"
)

var (
	buildCommit     string
	defaultTemplate = `<strong>{{ uppercasefirst build.status }}</strong> <a href="{{ system.link_url }}/{{ repo.owner }}/{{ repo.name }}/{{ build.number }}">{{ repo.owner }}/{{ repo.name }}#{{ truncate build.commit 8 }}</a> ({{ build.branch }}) by {{ build.author }} in {{ duration build.started_at build.finished_at }} </br> - {{ build.message }}`
	defaultCardTitleTemplate = `{{ build.status }}`
	defaultCardTemplate = `<strong>{{ repo.name }}</strong> ({{ build.branch }}) by {{ build.author }} in {{ duration build.started_at build.finished_at }} <a href="{{ build.link_url }}">{{ truncate build.commit 8 }}</a> - <i>{{ build.message }}</i>`
	defaultCardIcon = "http://readme.drone.io/logos/downstream.svg"
)

func main() {
	fmt.Printf("Drone HipChat Plugin built from %s\n", buildCommit)

	system := drone.System{}
	repo := drone.Repo{}
	build := drone.Build{}
	vargs := Params{}

	plugin.Param("system", &system)
	plugin.Param("repo", &repo)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	if len(vargs.Template) == 0 {
		vargs.Template = defaultTemplate
	}

	message := &Message{
		From:    vargs.From,
		Notify:  vargs.Notify,
		Color:   Color(&build),
		Message: BuildTemplate(
			&system,
			&repo,
			&build,
			vargs.Template,
		),
	}

	if vargs.UseCard {

		if len(vargs.CardTitleTemplate) == 0 {
			vargs.CardTitleTemplate = defaultCardTitleTemplate
		}

		if len(vargs.CardIcon) == 0 {
			vargs.CardIcon = defaultCardIcon
		}

		if len(vargs.CardTemplate) == 0 {
			vargs.CardTemplate = defaultCardTemplate
		}

		message.Card = &Card{
			ID:    build.Commit,
			Style: "link",
			Icon:  vargs.CardIcon,
			Title: BuildTemplate(
				&system,
				&repo,
				&build,
				vargs.CardTitleTemplate,
			),
			URL: BuildTemplate(
				&system,
				&repo,
				&build,
				"{{ system.link_url }}/{{ repo.owner }}/{{ repo.name }}/{{ build.number }}",
			),
			Description: Description{
				Format: "html",
				Value:  BuildTemplate(
					&system,
					&repo,
					&build,
					vargs.CardTemplate,
				),
			},
		}
	}

	client := NewClient(
		vargs.URL,
		vargs.Room.String(),
		vargs.Token,
	)

	if err := client.Send(message); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
}

// BuildMessage renders the HipChat message from a template.
func BuildTemplate(system *drone.System, repo *drone.Repo, build *drone.Build, tmpl string) string {

	payload := &drone.Payload{
		System: system,
		Repo:   repo,
		Build:  build,
	}

	msg, err := template.RenderTrim(tmpl, payload)
	if err != nil {
		return err.Error()
	}

	return msg
}

// Color determins the notfication color based upon the current build status.
func Color(build *drone.Build) string {
	switch build.Status {
	case drone.StatusSuccess:
		return "green"
	case drone.StatusFailure, drone.StatusError, drone.StatusKilled:
		return "red"
	default:
		return "yellow"
	}
}
