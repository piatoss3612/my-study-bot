package study

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrAdminNotFound   = errors.New("관리자 정보를 찾을 수 없습니다.")
	ErrUserNotFound    = errors.New("사용자 정보를 찾을 수 없습니다.")
	ErrChannelNotFound = errors.New("채널 정보를 찾을 수 없습니다.")
	ErrRequiredArgs    = errors.New("필수 인자가 없습니다.")
	ErrInvalidArgs     = errors.New("인자가 올바르지 않습니다.")
	ErrStudyNotFound   = errors.New("스터디 정보를 찾을 수 없습니다.")
	ErrMemberNotFound  = errors.New("스터디 멤버 정보를 찾을 수 없습니다.")
)

func errorInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
		},
	})
}

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
