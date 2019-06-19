package main

import (
	"context"
	"strconv"
	"strings"

	gh "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	open              = "open"
	approved          = "APPROVED"
	numberOfApprovals = 2
)

type GH struct {
	client *gh.Client
	url    string
}

func NewGitHubPlugin(url, token string) (*GH, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	c, err := gh.NewEnterpriseClient(url, url, tc)
	if err != nil {
		return nil, errors.Wrap(err, "attempt to connect to github failed")
	}

	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("creating a new github client...")

	return &GH{
		client: c,
		url:    url,
	}, nil
}

func (g *GH) ListActions() map[string]ConversationFn {
	return map[string]ConversationFn{
		"list-prs":                   g.ListPullRequests,
		"list-prs-without-approvals": g.ListPullRequestsWithoutApprovals,
	}
}

func (g *GH) ListPullRequests(ctx context.Context, c *Conversation) *ConversationResponse {

	logrus.WithFields(logrus.Fields{
		"owner": c.Params["org"],
		"repo":  c.Params["repository"],
		"state": c.Params["state"],
	}).Debug("listing pull requests")

	prs, _, err := g.client.PullRequests.List(ctx, c.Params["org"], c.Params["repository"], &gh.PullRequestListOptions{State: c.Params["state"]})
	if err != nil {
		logrus.WithField("conversation", c).Error(err)

		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	writer, err := NewTemplateWriter(c.ResponseTmpl)
	if err != nil {
		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	var resp []string
	for _, pr := range prs {
		resp = append(resp, writer.Write(pr))
	}

	return &ConversationResponse{
		Channel:  c.Channel,
		Text:     strings.Join(resp, "\n"),
		ParentID: c.ID,
	}
}

func (g *GH) ListPullRequestsWithoutApprovals(ctx context.Context, c *Conversation) *ConversationResponse {
	logrus.WithFields(logrus.Fields{
		"owner": c.Params["org"],
		"repo":  c.Params["repository"],
		"qtd":   c.Params["qtd"],
	}).Debug("listing pull requests without approvals")

	numberOfApprovals, err := strconv.Atoi(c.Params["qtd"])
	if err != nil {
		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	prs, _, err := g.client.PullRequests.List(ctx, c.Params["org"], c.Params["repository"], &gh.PullRequestListOptions{State: open})
	if err != nil {
		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	ret := make([]*gh.PullRequest, 0)
	for _, pr := range prs {
		rev, _, err := g.client.PullRequests.ListReviews(ctx, c.Params["org"], c.Params["repository"], pr.GetNumber(), nil)
		if err != nil {
			return &ConversationResponse{
				Channel:  c.Channel,
				Text:     err.Error(),
				ParentID: c.ID,
			}
		}
		count := 0
		for _, r := range rev {
			if *r.State == approved {
				count++
				if count == numberOfApprovals {
					break
				}
			}
		}
		if count < 2 {
			ret = append(ret, pr)
		}
	}

	writer, err := NewTemplateWriter(c.ResponseTmpl)
	if err != nil {
		return &ConversationResponse{
			Channel:  c.Channel,
			Text:     err.Error(),
			ParentID: c.ID,
		}
	}

	var resp []string
	for _, pr := range prs {
		resp = append(resp, writer.Write(pr))
	}

	return &ConversationResponse{
		Channel:  c.Channel,
		Text:     strings.Join(resp, "\n"),
		ParentID: c.ID,
	}
}
