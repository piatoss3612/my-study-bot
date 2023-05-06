package main

import (
	"sync"
	"time"
)

type StudyStage uint8

const (
	StudyStageNone               StudyStage = 0
	StudyStageWait               StudyStage = 1
	StudyStageStartRegistration  StudyStage = 2
	StudyStageFinishRegistration StudyStage = 3
	StudyStageStartSubmission    StudyStage = 4
	StudyStageFinishSubmission   StudyStage = 5
	StudyStageStartPresentation  StudyStage = 6
	StudyStageFinishPresentation StudyStage = 7
	StudyStageStartReview        StudyStage = 8
	StudyStageFinishReview       StudyStage = 9
)

func (s StudyStage) String() string {
	switch s {
	case StudyStageWait:
		return "다음 회차 대기"
	case StudyStageStartRegistration, StudyStageFinishRegistration:
		return "발표자 등록"
	case StudyStageStartSubmission, StudyStageFinishSubmission:
		return "발표자료 제출"
	case StudyStageStartPresentation, StudyStageFinishPresentation:
		return "발표"
	case StudyStageStartReview, StudyStageFinishReview:
		return "리뷰 및 피드백"
	default:
		return "몰?루"
	}
}

func (s StudyStage) IsNone() bool {
	return s == StudyStageNone
}

func (s StudyStage) IsWait() bool {
	return s == StudyStageWait
}

func (s StudyStage) IsRegister() bool {
	return s == StudyStageStartRegistration
}

func (s StudyStage) IsSubmit() bool {
	return s == StudyStageStartSubmission
}

func (s StudyStage) IsPresent() bool {
	return s == StudyStageStartPresentation
}

func (s StudyStage) IsReview() bool {
	return s == StudyStageStartReview
}

type StudyManager struct {
	GuildID         string
	NoticeChannelID string

	ManagerID     string
	SubManagerIDs []string

	OnGoingStudyID string
	StudyStage     StudyStage

	mtx *sync.Mutex
}

func NewStudyManager(guildID string, ManagerID string) *StudyManager {
	return &StudyManager{
		GuildID:         guildID,
		NoticeChannelID: "",
		ManagerID:       ManagerID,
		SubManagerIDs:   []string{},
		OnGoingStudyID:  "",
		StudyStage:      StudyStageNone,
		mtx:             &sync.Mutex{},
	}
}

func (s *StudyManager) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.NoticeChannelID = channelID
}

func (s *StudyManager) IsManager(userID string) bool {
	return s.ManagerID == userID
}

func (s *StudyManager) AddSubManagerID(userID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.SubManagerIDs = append(s.SubManagerIDs, userID)
}

func (s *StudyManager) RemoveSubManagerID(userID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	for i, v := range s.SubManagerIDs {
		if v == userID {
			s.SubManagerIDs = append(s.SubManagerIDs[:i], s.SubManagerIDs[i+1:]...)
		}
	}
}

func (s *StudyManager) IsSubManager(userID string) bool {
	for _, v := range s.SubManagerIDs {
		if v == userID {
			return true
		}
	}
	return false
}

func (s *StudyManager) SetOnGoingStudyID(studyID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.OnGoingStudyID = studyID
}

func (s *StudyManager) SetStudyStage(state StudyStage) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.StudyStage = state
}

type Member struct {
	Name       string
	Registered bool

	Subject    string
	ContentURL string
	Completed  bool

	mtx *sync.Mutex
}

func NewMember(name string) Member {
	return Member{
		Name:       name,
		Registered: false,
		Subject:    "",
		ContentURL: "",
		Completed:  false,
		mtx:        &sync.Mutex{},
	}
}

func (m *Member) SetRegistered(registered bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()
	m.Registered = registered
}

func (m *Member) SetSubject(subject string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()
	m.Subject = subject
}

func (m *Member) SetContentURL(contentURL string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()
	m.ContentURL = contentURL
}

func (m *Member) SetCompleted(completed bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()
	m.Completed = completed
}

type Study struct {
	ID      string
	GuildID string

	Title     string
	Members   map[string]Member
	CreatedAt time.Time
	UpdatedAt time.Time

	mtx *sync.Mutex
}

func NewStudy(guildID, title string) *Study {
	return &Study{
		ID:        "",
		GuildID:   guildID,
		Title:     title,
		Members:   map[string]Member{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		mtx:       &sync.Mutex{},
	}
}

func (s *Study) SetID(id string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.ID = id
}

func (s *Study) SetTitle(title string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.Title = title
}

func (s *Study) SetMember(memberID string, member Member) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.Members[memberID] = member
}

func (s *Study) RemoveMember(memberID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	delete(s.Members, memberID)
}

func (s *Study) GetMember(memberID string) (Member, bool) {
	member, ok := s.Members[memberID]
	return member, ok
}

func (s *Study) GetMembers() []Member {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	members := []Member{}
	for _, v := range s.Members {
		members = append(members, v)
	}
	return members
}

func (s *Study) SetUpdatedAt(updatedAt time.Time) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.UpdatedAt = updatedAt
}