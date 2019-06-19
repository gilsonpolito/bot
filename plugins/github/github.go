package github

import (
	"context"

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
}

func New(url, token string) (*GH, error) {
	c, err := newClient(url, token)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("creating a new github client...")

	return &GH{
		client: c,
	}, nil
}

func newClient(url, token string) (*gh.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cli, err := gh.NewEnterpriseClient(url, url, tc)
	if err != nil {
		return nil, errors.Wrap(err, "attempt to connect to github failed")
	}
	return cli, nil
}

func (g *GH) ListPullRequests(ctx context.Context, owner, repo, state string) ([]*gh.PullRequest, error) {
	prs, _, err := g.client.PullRequests.List(ctx, owner, repo, &gh.PullRequestListOptions{State: state})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (g *GH) ListPullRequestsWithoutTwoApproved(ctx context.Context, owner, repo, state string) ([]*gh.PullRequest, error) {
	prs, err := g.ListPullRequests(ctx, owner, repo, open)
	if err != nil {
		return nil, err
	}

	ret := make([]*gh.PullRequest, 0)
	for _, pr := range prs {
		rev, _, err := g.client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), nil)
		if err != nil {
			return nil, err
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

	return ret, nil
}
