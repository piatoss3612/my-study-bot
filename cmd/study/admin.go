package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var adminCmd = discordgo.ApplicationCommand{
	Name:        "매니저",
	Description: "스터디 관리 명령어입니다. 매니저만 사용할 수 있습니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "명령어",
			Description: "사용할 명령어를 선택해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "스터디 생성",
					Value: "create-study",
				},
				{
					Name:  "발표자 등록 마감",
					Value: "close-registration",
				},
				{
					Name:  "발표자 등록 취소",
					Value: "cancel-registration",
				},
				{
					Name:  "발표자료 제출 시작",
					Value: "start-submission",
				},
				{
					Name:  "발표자료 제출 마감",
					Value: "close-submission",
				},
				{
					Name:  "발표 시작",
					Value: "start-presentation",
				},
				{
					Name:  "발표 종료",
					Value: "end-presentation",
				},
				{
					Name:  "피드백 시작",
					Value: "start-feedback",
				},
				{
					Name:  "피드백 종료",
					Value: "end-feedback",
				},
				{
					Name:  "스터디 종료",
					Value: "end-study",
				},
				{
					Name:  "공지 채널 설정",
					Value: "set-notice-channel",
				},
			},
			Required: true,
		},
		{
			Name:        "제목",
			Description: "스터디 제목을 입력해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
		},
		{
			Name:        "스터디원",
			Description: "스터디원을 선택해주세요.",
			Type:        discordgo.ApplicationCommandOptionUser,
		},
		{
			Name:        "채널",
			Description: "채널을 선택해주세요.",
			Type:        discordgo.ApplicationCommandOptionChannel,
		},
	},
}

func (b *StudyBot) adminHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil || user.ID != b.svc.GetManagerID() {
		return
	}

	options := i.ApplicationCommandData().Options

	cmd := options[0].StringValue()

	var title string
	var u *discordgo.User
	var ch *discordgo.Channel

	for _, o := range options[1:] {
		switch o.Name {
		case "제목":
			title = o.StringValue()
		case "스터디원":
			u = o.UserValue(s)
		case "채널":
			ch = o.ChannelValue(s)
		}
	}

	var err error

	switch cmd {
	case "create-study":
		err = b.createStudyHandler(s, i, title)
	case "close-registration":
		err = b.closeRegistrationHandler(s, i)
	case "cancel-registration":
		err = b.cancelRegistrationHandler(s, i, u)
	case "start-submission":
		err = b.startSubmissionHandler(s, i)
	case "close-submission":
		err = b.closeSubmissionHandler(s, i)
	case "start-presentation":
		err = b.startPresentationHandler(s, i)
	case "end-presentation":
		err = b.endPresentationHandler(s, i)
	case "start-feedback":
		err = b.startFeedbackHandler(s, i)
	case "end-feedback":
		err = b.endFeedbackHandler(s, i)
	case "end-study":
		err = b.endStudyHandler(s, i)
	case "set-notice-channel":
		err = b.setNoticeChannelHandler(s, i, ch)
	default:
		err = errors.New("invalid command")
	}

	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}

	_ = s.UpdateGameStatus(0, b.svc.GetCurrentStudyStage().String())
}

func (b *StudyBot) createStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate, title string) error {
	guildID := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guildID != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	// check if title is empty
	if title == "" {
		return errors.New("title is empty")
	}

	// get all members in the guild
	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		return err
	}

	// get all id of non-bot members
	var memberIDs []string

	for _, m := range members {
		if m.User == nil || m.User.Bot {
			continue
		}

		memberIDs = append(memberIDs, m.User.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// create a study
	err = b.svc.CreateStudy(ctx, i.Member.User.ID, title, memberIDs)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "스터디 생성", fmt.Sprintf("**<%s>**가 생성되었습니다.", title)))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디가 생성되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) closeRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guildID != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// close registration
	err := b.svc.FinishRegistration(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자 등록 마감", "발표자 등록이 마감되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자 등록이 마감되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) cancelRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) error {
	guild := b.svc.GetGuildID()

	// check if user is not nil
	if u == nil {
		return errors.New("user is nil")
	}

	if u.Bot {
		return errors.New("bot cannot cancel registration")
	}

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// cancel registration
	err := b.svc.ChangeMemberRegistration(ctx, i.GuildID, u.ID, "", "", false)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s>님의 발표자 등록이 취소되었습니다.", u.ID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) startSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start submission
	err := b.svc.StartSubmission(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자료 제출 시작", "발표자료 제출이 시작되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자료 제출이 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) closeSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// close submission
	err := b.svc.FinishSubmission(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자료 제출 마감", "발표자료 제출이 마감되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자료 제출이 마감되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) startPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start presentation
	err := b.svc.StartPresentation(ctx, i.Member.User.ID)

	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표 시작", "발표가 시작되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표가 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) endPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// end presentation
	err := b.svc.FinishPresentation(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표 종료", "발표가 종료되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표가 종료되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) startFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start feedback
	err := b.svc.StartReview(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "피드백 시작", "피드백이 시작되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "피드백이 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) endFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// end feedback
	err := b.svc.FinishReview(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "피드백 종료", "피드백이 종료되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "피드백이 종료되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) endStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// end study
	err := b.svc.FinishStudy(ctx, i.Member.User.ID)
	if err != nil {
		return err
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "스터디 종료", "스터디가 종료되었습니다."))
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디가 종료되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) setNoticeChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
	guild := b.svc.GetGuildID()

	// check if the channel is nil
	if ch == nil {
		return errors.New("channel is nil")
	}

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		return errors.New("guild id mismatch with the bot's guild id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// set notice channel
	err := b.svc.SetNoticeChannelID(ctx, i.Member.User.ID, ch.ID)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("공지 채널이 %s로 설정되었습니다.", ch.Mention()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
