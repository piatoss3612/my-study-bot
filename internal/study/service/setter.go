package study

import (
	"context"
	"errors"
	"fmt"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

// set notice channel id
func (svc *serviceImpl) SetNoticeChannelID(ctx context.Context, guildID, channelID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		s.SetNoticeChannelID(channelID)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set reflection channel id
func (svc *serviceImpl) SetReflectionChannelID(ctx context.Context, guildID, channelID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		s.SetReflectionChannelID(channelID)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set member registration
func (svc *serviceImpl) SetMemberRegistration(ctx context.Context, guildID, memberID, name, subject string, register bool) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	// transaction for changing member registration
	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		if s.CurrentStage != study.StageRegistrationOpened {
			return nil, errors.Join(study.ErrInvalidStage, errors.New("발표자 등록 및 등록 해지가 불가능한 단계입니다"))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			member = study.NewMember()
		}

		if register {
			// check if member is already registered
			if member.Registered {
				return nil, study.ErrAlreadyRegistered
			}
			member.SetName(name)
			member.SetSubject(subject)
		} else {
			// check if member is not registered
			if !member.Registered {
				return nil, study.ErrNotRegistered
			}
			member.SetName("")
			member.SetSubject("")
		}

		member.SetRegistered(register)

		// set updated member to study
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set member content
func (svc *serviceImpl) SetMemberContent(ctx context.Context, guildID, memberID, contentURL string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// check if study is in submission stage
		if s.CurrentStage != study.StageSubmissionOpened {
			return nil, errors.Join(study.ErrInvalidStage, errors.New("발표 자료 제출이 불가능합니다"))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			return nil, study.ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, study.ErrNotRegistered
		}

		// set content
		member.SetContentURL(contentURL)

		// set updated member to round
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set speaker attended
func (svc *serviceImpl) SetSpeakerAttended(ctx context.Context, guildID, memberID string, attended bool) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// check if presentation is started
		if s.CurrentStage < study.StagePresentationStarted {
			return nil, errors.Join(study.ErrInvalidStage, errors.New("발표자 출석 확인이 불가능합니다"))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			return nil, study.ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, study.ErrNotRegistered
		}

		// set attended
		member.SetAttended(attended)

		// set updated member to study
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set study content
func (svc *serviceImpl) SetStudyContent(ctx context.Context, guildID, content string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// check if presentation is finished
		if s.CurrentStage < study.StagePresentationFinished {
			return nil, errors.Join(study.ErrInvalidStage, errors.New("스터디 자료 링크 등록이 불가능합니다"))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// set content
		r.SetContentURL(content)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set reviewer
func (svc *serviceImpl) SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if reviewerID == revieweeID {
		return study.ErrReviewByYourself
	}

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// check if review is ongoing
		if s.CurrentStage != study.StageReviewOpened {
			return nil, errors.Join(study.ErrInvalidStage, errors.New("리뷰어 지정이 불가능합니다"))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// check if reviewer is member of ongoing study
		_, ok := r.GetMember(reviewerID)
		if !ok {
			return nil, errors.Join(study.ErrMemberNotFound, errors.New("스터디에 참여한 사용자만 리뷰 참여가 가능합니다"))
		}

		// check if reviewee is member of ongoing study
		reviewee, ok := r.GetMember(revieweeID)
		if !ok {
			return nil, errors.Join(study.ErrMemberNotFound, errors.New("리뷰 대상자는 스터디에 참여한 사용자여야 합니다"))
		}

		// check if reviewee is registered and attended presentation
		if !reviewee.Registered || !reviewee.Attended {
			return nil, errors.New("리뷰 대상자는 발표에 참여한 사용자여야 합니다")
		}

		// check if reviewer already reviewed
		if reviewee.IsReviewer(reviewerID) {
			return nil, errors.New("이미 리뷰를 작성하였습니다")
		}

		// set reviewer
		reviewee.SetReviewer(reviewerID)

		// set updated member to study
		r.SetMember(revieweeID, reviewee)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set sentReflection of member
func (svc *serviceImpl) SetSentReflection(ctx context.Context, guildID, memberID string) (string, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		if s.CurrentStage < study.StagePresentationFinished {
			return nil, errors.Join(study.ErrInvalidStage, fmt.Errorf("%s 단계에서는 회고를 작성할 수 없습니다", s.CurrentStage.String()))
		}

		// check if there is ongoing round
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		if s.ReflectionChannelID == "" {
			return nil, errors.New("회고 채널이 설정되지 않았습니다")
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// check if member exists
		m, ok := r.GetMember(memberID)
		if !ok {
			return nil, study.ErrMemberNotFound
		}

		// check if member is registered and attended presentation
		if !m.Registered || !m.Attended {
			return nil, errors.New("발표에 참여한 사용자만 회고를 작성할 수 있습니다")
		}

		// check if member already sent reflection
		if m.SentReflection {
			return nil, study.ErrAlreadySentReflection
		}

		// set sent reflection
		m.SetSentReflection(true)

		// set updated member to round
		r.SetMember(memberID, m)

		// update round
		_, err = svc.tx.UpdateRound(sc, *r)
		if err != nil {
			return nil, err
		}

		return s.ReflectionChannelID, nil
	}

	// execute transaction
	id, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return "", err
	}

	// return reflection channel id
	return id.(string), nil
}