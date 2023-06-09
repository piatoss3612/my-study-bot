package service

import (
	"github.com/piatoss3612/my-study-bot/internal/study"
)

func MoveStage(s *study.Study, r *study.Round, _ *UpdateParams) {
	next := s.CurrentStage.Next()

	if next == study.StageFinished {
		s.SetCurrentStage(study.StageWait)
		s.SetOngoingRoundID("")
		r.SetStage(next)
		return
	}

	s.SetCurrentStage(next)
	r.SetStage(next)
}

func UpdateManagerID(s *study.Study, _ *study.Round, params *UpdateParams) {
	s.SetManagerID(params.ManagerID)
}

func UpdateNoticeChannelID(s *study.Study, _ *study.Round, params *UpdateParams) {
	s.SetNoticeChannelID(params.ChannelID)
}

func UpdateReflectionChannelID(s *study.Study, _ *study.Round, params *UpdateParams) {
	s.SetReflectionChannelID(params.ChannelID)
}

func RegisterMember(_ *study.Study, r *study.Round, params *UpdateParams) {
	member, ok := r.GetMember(params.MemberID)
	if !ok {
		member = study.NewMember()
	}

	member.SetName(params.MemberName)
	member.SetSubject(params.Subject)
	member.SetRegistered(true)

	r.SetMember(params.MemberID, member)
}

func SubmitMemberContent(_ *study.Study, r *study.Round, params *UpdateParams) {
	member, _ := r.GetMember(params.MemberID)
	member.SetContentURL(params.ContentURL)

	r.SetMember(params.MemberID, member)
}

func CheckSpeakerAttendance(_ *study.Study, r *study.Round, params *UpdateParams) {
	member, _ := r.GetMember(params.MemberID)
	member.SetAttended(true)

	r.SetMember(params.MemberID, member)
}

func SubmitRoundContent(_ *study.Study, r *study.Round, params *UpdateParams) {
	r.SetContentURL(params.ContentURL)
}

func SetReviewer(_ *study.Study, r *study.Round, params *UpdateParams) {
	reviewee, _ := r.GetMember(params.RevieweeID)
	reviewee.SetReviewer(params.ReviewerID)

	r.SetMember(params.RevieweeID, reviewee)
}

func SetSentReflection(_ *study.Study, r *study.Round, params *UpdateParams) {
	member, _ := r.GetMember(params.MemberID)
	member.SetSentReflection(true)

	r.SetMember(params.MemberID, member)
}

func SetSpreadsheetURL(s *study.Study, _ *study.Round, params *UpdateParams) {
	s.SetSpreadsheetURL(params.ContentURL)
}
