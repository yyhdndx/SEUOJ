package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
	"seu-oj-backend/internal/repository"
)

var (
	ErrContestNotFound             = errors.New("contest not found")
	ErrContestForbidden            = errors.New("contest forbidden")
	ErrContestInvalidTimeRange     = errors.New("contest invalid time range")
	ErrContestInvalidProblemSet    = errors.New("contest invalid problem set")
	ErrContestRegistrationClosed   = errors.New("contest registration closed")
	ErrContestNotRegistered        = errors.New("contest not registered")
	ErrContestNotRunning           = errors.New("contest not running")
	ErrContestProblemNotFound      = errors.New("contest problem not found")
	ErrContestAnnouncementNotFound = errors.New("contest announcement not found")
)

type ContestService struct {
	db                  *gorm.DB
	problemRepo         *repository.ProblemRepository
	problemTestcaseRepo *repository.ProblemTestcaseRepository
}

type contestListRecord struct {
	Contest         model.Contest
	RegisteredCount int64
	ProblemCount    int64
}

type contestParticipantRow struct {
	UserID    uint64 `gorm:"column:user_id"`
	Username  string `gorm:"column:username"`
	StudentID string `gorm:"column:userid"`
}

type contestProblemCellState struct {
	problemID         uint64
	problemCode       string
	solved            bool
	wrongAttempts     int
	acceptedAt        *time.Time
	penaltyMinutes    int
	latestStatus      string
	submissionCount   int
	frozenSubmissions int
}

type contestRankEntry struct {
	UserID          uint64
	Username        string
	StudentID       string
	SolvedCount     int
	PenaltyMinutes  int
	LastAcceptedAt  *time.Time
	SubmissionCount int
	Cells           []contestProblemCellState
}

func NewContestService(db *gorm.DB, problemRepo *repository.ProblemRepository, problemTestcaseRepo *repository.ProblemTestcaseRepository) *ContestService {
	return &ContestService{
		db:                  db,
		problemRepo:         problemRepo,
		problemTestcaseRepo: problemTestcaseRepo,
	}
}

func (s *ContestService) List(page, pageSize int, keyword, status string) (*dto.ContestListResponse, error) {
	return s.list(page, pageSize, keyword, status, false)
}

func (s *ContestService) ListAdmin(role string, page, pageSize int, keyword, status string) (*dto.ContestListResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	return s.list(page, pageSize, keyword, status, true)
}

func (s *ContestService) list(page, pageSize int, keyword, status string, includePrivate bool) (*dto.ContestListResponse, error) {
	var contests []model.Contest
	var total int64
	query := s.db.Model(&model.Contest{})
	if !includePrivate {
		query = query.Where("is_public = ?", true)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		query = query.Where("title LIKE ?", "%"+trimmed+"%")
	}
	now := time.Now()
	switch strings.TrimSpace(status) {
	case "upcoming":
		query = query.Where("start_time > ?", now)
	case "running":
		query = query.Where("start_time <= ? AND end_time > ?", now, now)
	case "ended":
		query = query.Where("end_time <= ?", now)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Order("start_time DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&contests).Error; err != nil {
		return nil, err
	}
	items := make([]dto.ContestListItem, 0, len(contests))
	for _, contest := range contests {
		registeredCount, problemCount, err := s.countContestMeta(contest.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, dto.ContestListItem{
			ID:               contest.ID,
			Title:            contest.Title,
			RuleType:         contest.RuleType,
			Status:           contestStatus(contest),
			StartTime:        contest.StartTime,
			EndTime:          contest.EndTime,
			IsPublic:         contest.IsPublic,
			AllowPractice:    contest.AllowPractice,
			RanklistFreezeAt: contest.RanklistFreezeAt,
			RanklistFrozen:   contestRanklistFrozen(contest, false),
			RegisteredCount:  registeredCount,
			ProblemCount:     problemCount,
		})
	}
	return &dto.ContestListResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ContestService) GetByID(id uint64) (*dto.ContestDetailResponse, error) {
	contest, err := s.getContest(id)
	if err != nil {
		return nil, err
	}
	if !contest.IsPublic {
		return nil, ErrContestNotFound
	}
	return s.buildContestDetailResponse(contest, false)
}

func (s *ContestService) GetAdminByID(role string, id uint64) (*dto.ContestDetailResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	contest, err := s.getContest(id)
	if err != nil {
		return nil, err
	}
	return s.buildContestDetailResponse(contest, true)
}

func (s *ContestService) Create(userID uint64, role string, req dto.CreateContestRequest) (uint64, error) {
	if role != "admin" {
		return 0, ErrPermissionDenied
	}
	contest, problems, err := s.buildContestForWrite(userID, req)
	if err != nil {
		return 0, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&contest).Error; err != nil {
			return err
		}
		for i := range problems {
			problems[i].ContestID = contest.ID
		}
		return tx.Create(&problems).Error
	}); err != nil {
		return 0, err
	}
	return contest.ID, nil
}

func (s *ContestService) Update(role string, id uint64, req dto.CreateContestRequest) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	contest, err := s.getContest(id)
	if err != nil {
		return err
	}
	updatedContest, problems, err := s.buildContestForWrite(contest.CreatedBy, req)
	if err != nil {
		return err
	}
	contest.Title = updatedContest.Title
	contest.Description = updatedContest.Description
	contest.RuleType = updatedContest.RuleType
	contest.StartTime = updatedContest.StartTime
	contest.EndTime = updatedContest.EndTime
	contest.IsPublic = updatedContest.IsPublic
	contest.AllowPractice = updatedContest.AllowPractice
	contest.RanklistFreezeAt = updatedContest.RanklistFreezeAt

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(contest).Error; err != nil {
			return err
		}
		if err := tx.Where("contest_id = ?", id).Delete(&model.ContestProblem{}).Error; err != nil {
			return err
		}
		for i := range problems {
			problems[i].ContestID = id
		}
		return tx.Create(&problems).Error
	})
}

func (s *ContestService) Delete(role string, id uint64) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	if _, err := s.getContest(id); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("contest_id = ?", id).Delete(&model.ContestRegistration{}).Error; err != nil {
			return err
		}
		if err := tx.Where("contest_id = ?", id).Delete(&model.ContestProblem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Contest{}, id).Error
	})
}

func (s *ContestService) Register(userID uint64, role string, contestID uint64) error {
	contest, err := s.getContest(contestID)
	if err != nil {
		return err
	}
	if role != "admin" && !contest.IsPublic {
		return ErrContestForbidden
	}
	if contestStatus(*contest) == "ended" && !contest.AllowPractice {
		return ErrContestRegistrationClosed
	}
	registration := model.ContestRegistration{ContestID: contestID, UserID: userID}
	return s.db.Where("contest_id = ? AND user_id = ?", contestID, userID).FirstOrCreate(&registration).Error
}

func (s *ContestService) GetMe(userID uint64, role string, contestID uint64) (*dto.ContestMeResponse, error) {
	contest, err := s.getContest(contestID)
	if err != nil {
		return nil, err
	}
	if role != "admin" && !contest.IsPublic {
		return nil, ErrContestForbidden
	}
	status := contestStatus(*contest)
	practiceEnabled := contestPracticeEnabled(*contest)
	registered, err := s.isRegistered(contestID, userID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ContestMeResponse{
		ContestID:       contestID,
		Registered:      registered,
		ContestStatus:   status,
		CanRegister:     !registered && (status != "ended" || contest.AllowPractice),
		CanViewProblems: role == "admin" || (registered && status != "upcoming"),
		CanSubmit:       role == "admin" || (registered && (status == "running" || practiceEnabled)),
		PracticeEnabled: practiceEnabled,
	}
	return resp, nil
}

func (s *ContestService) ListProblems(userID uint64, role string, contestID uint64) (*dto.ContestProblemListResponse, error) {
	contest, err := s.getContest(contestID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureContestProblemViewAllowed(*contest, userID, role); err != nil {
		return nil, err
	}
	items, err := s.loadContestProblemItems(contestID)
	if err != nil {
		return nil, err
	}
	return &dto.ContestProblemListResponse{
		ContestID:     contest.ID,
		ContestTitle:  contest.Title,
		ContestStatus: contestStatus(*contest),
		List:          items,
	}, nil
}

func (s *ContestService) GetProblemDetail(userID uint64, role string, contestID, problemID uint64) (*dto.ProblemDetailResponse, error) {
	contest, err := s.getContest(contestID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureContestProblemViewAllowed(*contest, userID, role); err != nil {
		return nil, err
	}
	mapping, err := s.getContestProblem(contestID, problemID)
	if err != nil {
		return nil, err
	}
	problem, err := s.problemRepo.GetByID(mapping.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestProblemNotFound
		}
		return nil, err
	}
	testcases, err := s.problemTestcaseRepo.ListByProblemID(problem.ID)
	if err != nil {
		return nil, err
	}
	resp := toProblemDetailResponse(*problem, filterSampleTestcases(testcases), nil)
	return &resp, nil
}

func (s *ContestService) GetRanklist(id uint64) (*dto.ContestRanklistResponse, error) {
	contest, err := s.getContest(id)
	if err != nil {
		return nil, err
	}
	if !contest.IsPublic {
		return nil, ErrContestNotFound
	}
	return s.buildRanklist(contest, false)
}

func (s *ContestService) GetAdminRanklist(role string, id uint64) (*dto.ContestRanklistResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	contest, err := s.getContest(id)
	if err != nil {
		return nil, err
	}
	return s.buildRanklist(contest, true)
}

func (s *ContestService) ValidateSubmissionAccess(userID uint64, role string, contestID, problemID uint64) error {
	contest, err := s.getContest(contestID)
	if err != nil {
		return err
	}
	if role == "admin" {
		_, err := s.getContestProblem(contestID, problemID)
		return err
	}
	if !contest.IsPublic {
		return ErrContestForbidden
	}
	status := contestStatus(*contest)
	if status != "running" && !(status == "ended" && contest.AllowPractice) {
		return ErrContestNotRunning
	}
	registered, err := s.isRegistered(contestID, userID)
	if err != nil {
		return err
	}
	if !registered {
		return ErrContestNotRegistered
	}
	_, err = s.getContestProblem(contestID, problemID)
	return err
}

func (s *ContestService) ValidateProblemAccess(userID uint64, role string, contestID, problemID uint64) error {
	contest, err := s.getContest(contestID)
	if err != nil {
		return err
	}
	if err := s.ensureContestProblemViewAllowed(*contest, userID, role); err != nil {
		return err
	}
	_, err = s.getContestProblem(contestID, problemID)
	return err
}

func (s *ContestService) ensureContestProblemViewAllowed(contest model.Contest, userID uint64, role string) error {
	if role == "admin" {
		return nil
	}
	if !contest.IsPublic {
		return ErrContestForbidden
	}
	if contestStatus(contest) == "upcoming" {
		return ErrContestForbidden
	}
	registered, err := s.isRegistered(contest.ID, userID)
	if err != nil {
		return err
	}
	if !registered {
		return ErrContestNotRegistered
	}
	return nil
}

func (s *ContestService) buildContestForWrite(createdBy uint64, req dto.CreateContestRequest) (model.Contest, []model.ContestProblem, error) {
	if !req.EndTime.After(req.StartTime) {
		return model.Contest{}, nil, ErrContestInvalidTimeRange
	}
	if len(req.Problems) == 0 {
		return model.Contest{}, nil, ErrContestInvalidProblemSet
	}
	problemMappings := make([]model.ContestProblem, 0, len(req.Problems))
	seenProblemIDs := map[uint64]struct{}{}
	seenCodes := map[string]struct{}{}
	for index, item := range req.Problems {
		if _, exists := seenProblemIDs[item.ProblemID]; exists {
			return model.Contest{}, nil, fmt.Errorf("%w: duplicated problem id %d", ErrContestInvalidProblemSet, item.ProblemID)
		}
		if _, err := s.problemRepo.GetByID(item.ProblemID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.Contest{}, nil, fmt.Errorf("%w: problem %d not found", ErrContestInvalidProblemSet, item.ProblemID)
			}
			return model.Contest{}, nil, err
		}
		code := strings.ToUpper(strings.TrimSpace(item.ProblemCode))
		if code == "" {
			code = defaultContestProblemCode(index)
		}
		if _, exists := seenCodes[code]; exists {
			return model.Contest{}, nil, fmt.Errorf("%w: duplicated problem code %s", ErrContestInvalidProblemSet, code)
		}
		displayOrder := item.DisplayOrder
		if displayOrder <= 0 {
			displayOrder = index + 1
		}
		problemMappings = append(problemMappings, model.ContestProblem{
			ProblemID:    item.ProblemID,
			ProblemCode:  code,
			DisplayOrder: displayOrder,
		})
		seenProblemIDs[item.ProblemID] = struct{}{}
		seenCodes[code] = struct{}{}
	}
	sort.Slice(problemMappings, func(i, j int) bool {
		if problemMappings[i].DisplayOrder == problemMappings[j].DisplayOrder {
			return problemMappings[i].ProblemCode < problemMappings[j].ProblemCode
		}
		return problemMappings[i].DisplayOrder < problemMappings[j].DisplayOrder
	})
	contest := model.Contest{
		Title:            strings.TrimSpace(req.Title),
		Description:      req.Description,
		RuleType:         req.RuleType,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		IsPublic:         req.IsPublic,
		AllowPractice:    req.AllowPractice,
		RanklistFreezeAt: req.RanklistFreezeAt,
		CreatedBy:        createdBy,
	}
	return contest, problemMappings, nil
}

func (s *ContestService) buildContestDetailResponse(contest *model.Contest, includeProblems bool) (*dto.ContestDetailResponse, error) {
	registeredCount, problemCount, err := s.countContestMeta(contest.ID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ContestDetailResponse{
		ID:               contest.ID,
		Title:            contest.Title,
		Description:      contest.Description,
		RuleType:         contest.RuleType,
		Status:           contestStatus(*contest),
		StartTime:        contest.StartTime,
		EndTime:          contest.EndTime,
		IsPublic:         contest.IsPublic,
		AllowPractice:    contest.AllowPractice,
		RanklistFreezeAt: contest.RanklistFreezeAt,
		RanklistFrozen:   contestRanklistFrozen(*contest, false),
		CreatedBy:        contest.CreatedBy,
		CreatedAt:        contest.CreatedAt,
		UpdatedAt:        contest.UpdatedAt,
		RegisteredCount:  registeredCount,
		ProblemCount:     problemCount,
	}
	if includeProblems {
		items, err := s.loadContestProblemItems(contest.ID)
		if err != nil {
			return nil, err
		}
		resp.Problems = items
	}
	return resp, nil
}

func (s *ContestService) ListAnnouncements(contestID uint64, page, pageSize int) (*dto.ContestAnnouncementListResponse, error) {
	contest, err := s.getContest(contestID)
	if err != nil {
		return nil, err
	}
	if !contest.IsPublic {
		return nil, ErrContestNotFound
	}
	return s.listAnnouncements(contestID, page, pageSize)
}

func (s *ContestService) GetAnnouncement(contestID, announcementID uint64) (*dto.ContestAnnouncementItem, error) {
	contest, err := s.getContest(contestID)
	if err != nil {
		return nil, err
	}
	if !contest.IsPublic {
		return nil, ErrContestNotFound
	}
	return s.getAnnouncementItem(contestID, announcementID)
}

func (s *ContestService) ListAdminAnnouncements(role string, contestID uint64, page, pageSize int) (*dto.ContestAnnouncementListResponse, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	if _, err := s.getContest(contestID); err != nil {
		return nil, err
	}
	return s.listAnnouncements(contestID, page, pageSize)
}

func (s *ContestService) GetAdminAnnouncement(role string, contestID, announcementID uint64) (*dto.ContestAnnouncementItem, error) {
	if role != "admin" {
		return nil, ErrPermissionDenied
	}
	if _, err := s.getContest(contestID); err != nil {
		return nil, err
	}
	return s.getAnnouncementItem(contestID, announcementID)
}

func (s *ContestService) CreateAnnouncement(userID uint64, role string, contestID uint64, req dto.CreateContestAnnouncementRequest) (uint64, error) {
	if role != "admin" {
		return 0, ErrPermissionDenied
	}
	if _, err := s.getContest(contestID); err != nil {
		return 0, err
	}
	item := model.ContestAnnouncement{
		ContestID: contestID,
		Title:     strings.TrimSpace(req.Title),
		Content:   req.Content,
		IsPinned:  req.IsPinned,
		CreatedBy: userID,
	}
	if err := s.db.Create(&item).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *ContestService) UpdateAnnouncement(role string, contestID, announcementID uint64, req dto.CreateContestAnnouncementRequest) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	if _, err := s.getContest(contestID); err != nil {
		return err
	}
	var item model.ContestAnnouncement
	if err := s.db.Where("contest_id = ? AND id = ?", contestID, announcementID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContestAnnouncementNotFound
		}
		return err
	}
	item.Title = strings.TrimSpace(req.Title)
	item.Content = req.Content
	item.IsPinned = req.IsPinned
	return s.db.Save(&item).Error
}

func (s *ContestService) DeleteAnnouncement(role string, contestID, announcementID uint64) error {
	if role != "admin" {
		return ErrPermissionDenied
	}
	if _, err := s.getContest(contestID); err != nil {
		return err
	}
	result := s.db.Where("contest_id = ? AND id = ?", contestID, announcementID).Delete(&model.ContestAnnouncement{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrContestAnnouncementNotFound
	}
	return nil
}

func (s *ContestService) listAnnouncements(contestID uint64, page, pageSize int) (*dto.ContestAnnouncementListResponse, error) {
	var list []model.ContestAnnouncement
	var total int64
	query := s.db.Model(&model.ContestAnnouncement{}).Where("contest_id = ?", contestID)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("is_pinned DESC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]dto.ContestAnnouncementItem, 0, len(list))
	for _, item := range list {
		items = append(items, dto.ContestAnnouncementItem{
			ID: item.ID, ContestID: item.ContestID, Title: item.Title, Content: item.Content, IsPinned: item.IsPinned, CreatedBy: item.CreatedBy, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	return &dto.ContestAnnouncementListResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ContestService) getAnnouncementItem(contestID, announcementID uint64) (*dto.ContestAnnouncementItem, error) {
	var item model.ContestAnnouncement
	if err := s.db.Where("contest_id = ? AND id = ?", contestID, announcementID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestAnnouncementNotFound
		}
		return nil, err
	}
	resp := dto.ContestAnnouncementItem{
		ID: item.ID, ContestID: item.ContestID, Title: item.Title, Content: item.Content, IsPinned: item.IsPinned, CreatedBy: item.CreatedBy, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
	return &resp, nil
}

func (s *ContestService) loadContestProblemItems(contestID uint64) ([]dto.ContestProblemItem, error) {
	var mappings []model.ContestProblem
	if err := s.db.Where("contest_id = ?", contestID).Order("display_order ASC, id ASC").Find(&mappings).Error; err != nil {
		return nil, err
	}
	items := make([]dto.ContestProblemItem, 0, len(mappings))
	for _, mapping := range mappings {
		problem, err := s.problemRepo.GetByID(mapping.ProblemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		items = append(items, dto.ContestProblemItem{
			ID:            mapping.ID,
			ProblemID:     mapping.ProblemID,
			ProblemCode:   mapping.ProblemCode,
			DisplayOrder:  mapping.DisplayOrder,
			Title:         problem.Title,
			JudgeMode:     problem.JudgeMode,
			TimeLimitMS:   problem.TimeLimitMS,
			MemoryLimitMB: problem.MemoryLimitMB,
		})
	}
	return items, nil
}

func (s *ContestService) buildRanklist(contest *model.Contest, includeFrozen bool) (*dto.ContestRanklistResponse, error) {
	problemItems, err := s.loadContestProblemItems(contest.ID)
	if err != nil {
		return nil, err
	}
	problemBriefs := make([]dto.ContestRanklistProblem, 0, len(problemItems))
	problemIndex := map[uint64]int{}
	for index, item := range problemItems {
		problemBriefs = append(problemBriefs, dto.ContestRanklistProblem{
			ProblemID:    item.ProblemID,
			ProblemCode:  item.ProblemCode,
			DisplayOrder: item.DisplayOrder,
		})
		problemIndex[item.ProblemID] = index
	}

	participants, err := s.loadContestParticipants(contest.ID)
	if err != nil {
		return nil, err
	}
	ranklistFrozen := contestRanklistFrozen(*contest, includeFrozen)
	freezeAt := contest.RanklistFreezeAt
	entries := make([]contestRankEntry, 0, len(participants))
	entryIndex := map[uint64]int{}
	for _, participant := range participants {
		cells := make([]contestProblemCellState, len(problemBriefs))
		for index, item := range problemBriefs {
			cells[index] = contestProblemCellState{problemID: item.ProblemID, problemCode: item.ProblemCode}
		}
		entries = append(entries, contestRankEntry{
			UserID:    participant.UserID,
			Username:  participant.Username,
			StudentID: participant.StudentID,
			Cells:     cells,
		})
		entryIndex[participant.UserID] = len(entries) - 1
	}

	var submissions []model.Submission
	if err := s.db.Where("contest_id = ?", contest.ID).Order("created_at ASC, id ASC").Find(&submissions).Error; err != nil {
		return nil, err
	}
	for _, submission := range submissions {
		if submission.IsPractice {
			continue
		}
		if ranklistFrozen && freezeAt != nil && submission.CreatedAt.After(*freezeAt) {
			entryPos, ok := entryIndex[submission.UserID]
			if !ok {
				continue
			}
			cellPos, ok := problemIndex[submission.ProblemID]
			if !ok {
				continue
			}
			entries[entryPos].Cells[cellPos].frozenSubmissions++
			continue
		}
		entryPos, ok := entryIndex[submission.UserID]
		if !ok {
			continue
		}
		cellPos, ok := problemIndex[submission.ProblemID]
		if !ok {
			continue
		}
		entry := &entries[entryPos]
		cell := &entry.Cells[cellPos]
		cell.submissionCount++
		entry.SubmissionCount++
		cell.latestStatus = submission.Status
		if cell.solved {
			continue
		}
		switch submission.Status {
		case "Accepted":
			cell.solved = true
			acceptedAt := submission.CreatedAt
			cell.acceptedAt = &acceptedAt
			minutes := int(acceptedAt.Sub(contest.StartTime).Minutes())
			if minutes < 0 {
				minutes = 0
			}
			cell.penaltyMinutes = minutes + cell.wrongAttempts*20
			entry.SolvedCount++
			entry.PenaltyMinutes += cell.penaltyMinutes
			if entry.LastAcceptedAt == nil || acceptedAt.After(*entry.LastAcceptedAt) {
				entry.LastAcceptedAt = &acceptedAt
			}
		case "Wrong Answer", "Runtime Error", "Time Limit Exceeded", "Memory Limit Exceeded":
			cell.wrongAttempts++
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].SolvedCount != entries[j].SolvedCount {
			return entries[i].SolvedCount > entries[j].SolvedCount
		}
		if entries[i].PenaltyMinutes != entries[j].PenaltyMinutes {
			return entries[i].PenaltyMinutes < entries[j].PenaltyMinutes
		}
		iLast := entries[i].LastAcceptedAt
		jLast := entries[j].LastAcceptedAt
		switch {
		case iLast == nil && jLast != nil:
			return false
		case iLast != nil && jLast == nil:
			return true
		case iLast != nil && jLast != nil && !iLast.Equal(*jLast):
			return iLast.Before(*jLast)
		}
		if entries[i].Username != entries[j].Username {
			return entries[i].Username < entries[j].Username
		}
		return entries[i].UserID < entries[j].UserID
	})

	items := make([]dto.ContestRanklistItem, 0, len(entries))
	currentRank := 0
	for index, entry := range entries {
		if index == 0 || !sameContestRank(entries[index-1], entry) {
			currentRank = index + 1
		}
		cells := make([]dto.ContestRanklistCell, 0, len(entry.Cells))
		for _, cell := range entry.Cells {
			cells = append(cells, dto.ContestRanklistCell{
				ProblemID:         cell.problemID,
				ProblemCode:       cell.problemCode,
				Solved:            cell.solved,
				WrongAttempts:     cell.wrongAttempts,
				AcceptedAt:        cell.acceptedAt,
				PenaltyMinutes:    cell.penaltyMinutes,
				LatestStatus:      cell.latestStatus,
				SubmissionCount:   cell.submissionCount,
				FrozenSubmissions: cell.frozenSubmissions,
			})
		}
		items = append(items, dto.ContestRanklistItem{
			Rank:            currentRank,
			UserID:          entry.UserID,
			Username:        entry.Username,
			StudentID:       entry.StudentID,
			SolvedCount:     entry.SolvedCount,
			PenaltyMinutes:  entry.PenaltyMinutes,
			LastAcceptedAt:  entry.LastAcceptedAt,
			SubmissionCount: entry.SubmissionCount,
			Cells:           cells,
		})
	}

	return &dto.ContestRanklistResponse{
		ContestID:        contest.ID,
		ContestTitle:     contest.Title,
		ContestStatus:    contestStatus(*contest),
		RanklistFrozen:   ranklistFrozen,
		RanklistFreezeAt: contest.RanklistFreezeAt,
		Problems:         problemBriefs,
		List:             items,
	}, nil
}

func sameContestRank(prev, current contestRankEntry) bool {
	if prev.SolvedCount != current.SolvedCount || prev.PenaltyMinutes != current.PenaltyMinutes {
		return false
	}
	switch {
	case prev.LastAcceptedAt == nil && current.LastAcceptedAt == nil:
		return true
	case prev.LastAcceptedAt == nil || current.LastAcceptedAt == nil:
		return false
	default:
		return prev.LastAcceptedAt.Equal(*current.LastAcceptedAt)
	}
}

func (s *ContestService) loadContestParticipants(contestID uint64) ([]contestParticipantRow, error) {
	var rows []contestParticipantRow
	err := s.db.Table("contest_registrations AS cr").
		Select("cr.user_id, u.username, u.userid").
		Joins("JOIN users u ON u.id = cr.user_id").
		Where("cr.contest_id = ?", contestID).
		Order("cr.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (s *ContestService) countContestMeta(contestID uint64) (int64, int64, error) {
	var registeredCount int64
	var problemCount int64
	if err := s.db.Model(&model.ContestRegistration{}).Where("contest_id = ?", contestID).Count(&registeredCount).Error; err != nil {
		return 0, 0, err
	}
	if err := s.db.Model(&model.ContestProblem{}).Where("contest_id = ?", contestID).Count(&problemCount).Error; err != nil {
		return 0, 0, err
	}
	return registeredCount, problemCount, nil
}

func (s *ContestService) isRegistered(contestID, userID uint64) (bool, error) {
	var count int64
	if err := s.db.Model(&model.ContestRegistration{}).Where("contest_id = ? AND user_id = ?", contestID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *ContestService) getContest(id uint64) (*model.Contest, error) {
	var contest model.Contest
	if err := s.db.First(&contest, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestNotFound
		}
		return nil, err
	}
	return &contest, nil
}

func (s *ContestService) getContestProblem(contestID, problemID uint64) (*model.ContestProblem, error) {
	var mapping model.ContestProblem
	if err := s.db.Where("contest_id = ? AND problem_id = ?", contestID, problemID).First(&mapping).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestProblemNotFound
		}
		return nil, err
	}
	return &mapping, nil
}

func contestStatus(contest model.Contest) string {
	now := time.Now()
	switch {
	case now.Before(contest.StartTime):
		return "upcoming"
	case now.After(contest.EndTime):
		return "ended"
	default:
		return "running"
	}
}

func contestPracticeEnabled(contest model.Contest) bool {
	return contest.AllowPractice && contestStatus(contest) == "ended"
}

func contestRanklistFrozen(contest model.Contest, includeFrozen bool) bool {
	if includeFrozen || contestStatus(contest) != "running" || contest.RanklistFreezeAt == nil {
		return false
	}
	return !time.Now().Before(*contest.RanklistFreezeAt)
}

func defaultContestProblemCode(index int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if index < len(letters) {
		return string(letters[index])
	}
	return fmt.Sprintf("P%d", index+1)
}

