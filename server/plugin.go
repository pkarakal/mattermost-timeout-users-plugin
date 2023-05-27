package main

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/logr/v2"
	"github.com/mattermost/mattermost-server/server/v8/model"
	"github.com/mattermost/mattermost-server/server/v8/platform/shared/i18n"
	"github.com/mattermost/mattermost-server/server/v8/platform/shared/mlog"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/server/v8/plugin"
)

// TimeoutUsersPlugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type TimeoutUsersPlugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *TimeoutUsersPlugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	if !filterChannelWideMentions(post.Message) {
		return post, ""
	}
	conf := p.getConfiguration()
	if conf.Disabled || conf.UserTimeoutInSeconds == 0 {
		return post, ""
	}

	postCreated := time.Unix(post.CreateAt, 0)
	p.API.LogDebug(postCreated.String())
	dateSince := postCreated.Add(time.Second * time.Duration(conf.UserTimeoutInSeconds))

	channelTeam, err := p.API.GetChannel(post.ChannelId)
	if err != nil {
		p.API.LogError("Couldn't get details for current channel", mlog.Field{
			Type:   logr.ErrorType,
			String: err.Error(),
		})
		return nil, "Couldn't get details for current channel"
	}
	userChannels, err := p.API.GetChannelMembersForUser(channelTeam.TeamId, post.UserId, 0, 1000)
	if err != nil {
		p.API.LogError("Couldn't get user channel memberships", mlog.Field{
			Type:   logr.ErrorType,
			String: err.Error(),
		})
		return nil, "Couldn't get user channel memberships"
	}

	posts, postErr := getUserPosts(p, userChannels, &dateSince)
	if postErr != nil {
		return nil, postErr.Error()
	}

	userPosts := filterUserChannelMentions(posts, post.UserId)

	user, err := p.API.GetUser(post.UserId)
	if err != nil {
		p.API.LogError(fmt.Sprintf("Couldn't get user with id %s", post.UserId))
		return nil, fmt.Sprintf("Couldn't get user with id %s", post.UserId)
	}

	if len(userPosts) > 0 && len(userPosts) > p.configuration.ChannelMentionsThreshold {
		sortPostsByCreationDate(userPosts)
		mostRecent := userPosts[0]
		timeoutExpiring := time.UnixMilli(mostRecent.CreateAt).Add(time.Duration(p.configuration.UserTimeoutInSeconds) * time.Second)
		remainingTimeout := time.Until(timeoutExpiring)
		if remainingTimeout < 0 {
			return post, ""
		}
		err := p.createEphemeralPost(timeoutExpiring, post.UserId, post.ChannelId, user)
		if err != nil {
			p.API.LogError("Couldn't post ephemeral message for message rejection")
			return nil, "Couldn't post ephemeral message"
		}
		return nil, fmt.Sprintf("You have reached the threshold of channel wide mentions of %d. You will be able to post a new message in %fs", p.configuration.ChannelMentionsThreshold-1, remainingTimeout.Seconds())
	}

	// Otherwise, allow the post through.
	return post, ""
}

func sortPostsByCreationDate(posts []*model.Post) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreateAt > posts[j].CreateAt
	})
}

func filterUserChannelMentions(posts []*model.Post, userID string) []*model.Post {
	filteredPosts := make([]*model.Post, 0)
	for _, v := range posts {
		if v.UserId == userID && filterChannelWideMentions(v.Message) {
			filteredPosts = append(filteredPosts, v)
		}
	}
	return filteredPosts
}

func (p *TimeoutUsersPlugin) createEphemeralPost(expiring time.Time, user, channel string, sender *model.User) error {
	bundle, err := p.API.GetBundlePath()
	if err != nil {
		return errors.New("couldn't get bundle path")
	}
	locale, _ := i18n.GetTranslationFuncForDir(path.Join(bundle, "assets", "i18n"))
	T := locale(sender.Locale)
	post := model.Post{
		UserId:    user,
		ChannelId: channel,
		Message: fmt.Sprintf("%s\n%s",
			T("com.pkarakal.mattermost_timeout_users_plugin.timeout.title"),
			T("com.pkarakal.mattermost_timeout_users_plugin.timeout.message", map[string]any{"Threshold": p.configuration.ChannelMentionsThreshold, "Expiring": expiring.Format(time.RFC1123)})),
		Type: model.PostTypeEphemeral,
	}
	res := p.API.SendEphemeralPost(user, &post)
	if res == nil {
		return errors.New("couldn't post ephemeral post to channel")
	}
	return nil
}

func filterChannelWideMentions(message string) bool {
	if strings.Contains(message, "@channel") || strings.Contains(message, "@all") || strings.Contains(message, "@here") {
		return true
	}
	return false
}

func getUserPosts(p *TimeoutUsersPlugin, channels []*model.ChannelMember, dateSince *time.Time) ([]*model.Post, error) {
	posts := make([]*model.Post, 0)

	for _, channel := range channels {
		res, err := p.API.GetPostsSince(channel.ChannelId, dateSince.Unix())
		if err != nil {
			p.API.LogError("Couldn't get messages in channel", mlog.Field{
				Type:   logr.ErrorType,
				String: err.Error(),
			})
			return nil, fmt.Errorf("couldn't get messages in channel %s", channel.ChannelId)
		}
		for _, v := range res.Posts {
			posts = append(posts, v)
		}
		for {
			page := 0
			if !res.HasNext {
				break
			}
			res, err = p.API.GetPostsAfter(channel.ChannelId, res.NextPostId, page, 1000)
			if err != nil {
				p.API.LogError(fmt.Sprintf("Couldn't get posts in channel after %s", res.NextPostId))
				return nil, fmt.Errorf("couldn't get posts in channel after %s", res.NextPostId)
			}
			for _, v := range res.Posts {
				posts = append(posts, v)
			}
		}
	}
	return posts, nil
}
