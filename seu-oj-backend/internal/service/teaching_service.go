package service

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"seu-oj-backend/internal/dto"
	"seu-oj-backend/internal/model"
)

var (
	ErrPlaylistNotFound   = errors.New("playlist not found")
	ErrClassNotFound      = errors.New("class not found")
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrJoinCodeInvalid    = errors.New("invalid join code")
	ErrAlreadyJoinedClass = errors.New("already joined class")
	ErrPlaylistInUse      = errors.New("playlist is in use")
	ErrTeachingForbidden  = errors.New("teaching access denied")
)

type TeachingService struct {
	db *gorm.DB
}

type assignmentStatusRow struct {
	ProblemID uint64
	Status    string
	CreatedAt time.Time
}

func NewTeachingService(db *gorm.DB) *TeachingService {
	return &TeachingService{db: db}
}

func (s *TeachingService) ListPublicPlaylists(page, pageSize int, keyword string) (*dto.PlaylistListResponse, error) {
	var items []dto.PlaylistListItem
	var total int64
	query := s.db.Table("playlists p").
		Select("p.id, p.title, p.description, p.visibility, p.created_by, p.created_at, COUNT(pp.id) AS problem_count").
		Joins("LEFT JOIN playlist_problems pp ON pp.playlist_id = p.id").
		Where("p.visibility = ?", "public").
		Group("p.id")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("p.title LIKE ?", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Order("p.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, err
	}
	return &dto.PlaylistListResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *TeachingService) GetPlaylistDetail(requestUserID uint64, requestRole string, id uint64, teacherView bool) (*dto.PlaylistDetailResponse, error) {
	var playlist model.Playlist
	if err := s.db.First(&playlist, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	if !teacherView && playlist.Visibility != "public" && requestRole != "admin" && playlist.CreatedBy != requestUserID {
		return nil, ErrTeachingForbidden
	}
	if teacherView && requestRole != "admin" && playlist.CreatedBy != requestUserID {
		return nil, ErrTeachingForbidden
	}
	var problems []dto.PlaylistProblemItem
	if err := s.db.Table("playlist_problems pp").
		Select("pp.problem_id, pp.display_order, p.display_id, p.title").
		Joins("JOIN problems p ON p.id = pp.problem_id").
		Where("pp.playlist_id = ?", id).
		Order("pp.display_order ASC, pp.id ASC").
		Scan(&problems).Error; err != nil {
		return nil, err
	}
	return &dto.PlaylistDetailResponse{ID: playlist.ID, Title: playlist.Title, Description: playlist.Description, Visibility: playlist.Visibility, CreatedBy: playlist.CreatedBy, CreatedAt: playlist.CreatedAt, UpdatedAt: playlist.UpdatedAt, Problems: problems}, nil
}

func (s *TeachingService) ListTeacherPlaylists(userID uint64, role string, page, pageSize int, keyword string) (*dto.PlaylistListResponse, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	var items []dto.PlaylistListItem
	var total int64
	query := s.db.Table("playlists p").
		Select("p.id, p.title, p.description, p.visibility, p.created_by, p.created_at, COUNT(pp.id) AS problem_count").
		Joins("LEFT JOIN playlist_problems pp ON pp.playlist_id = p.id")
	if role != "admin" {
		query = query.Where("p.created_by = ?", userID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("p.title LIKE ?", "%"+keyword+"%")
	}
	query = query.Group("p.id")
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := query.Order("p.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, err
	}
	return &dto.PlaylistListResponse{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *TeachingService) CreatePlaylist(userID uint64, role string, req dto.CreatePlaylistRequest) (*dto.PlaylistDetailResponse, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	if err := s.validatePlaylistProblems(req.Problems); err != nil {
		return nil, err
	}
	playlist := model.Playlist{Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description), Visibility: req.Visibility, CreatedBy: userID}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&playlist).Error; err != nil {
			return err
		}
		return s.replacePlaylistProblems(tx, playlist.ID, req.Problems)
	}); err != nil {
		return nil, err
	}
	return s.GetPlaylistDetail(userID, role, playlist.ID, true)
}

func (s *TeachingService) UpdatePlaylist(userID uint64, role string, id uint64, req dto.UpdatePlaylistRequest) (*dto.PlaylistDetailResponse, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	if err := s.validatePlaylistProblems(req.Problems); err != nil {
		return nil, err
	}
	var playlist model.Playlist
	if err := s.db.First(&playlist, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	if role != "admin" && playlist.CreatedBy != userID {
		return nil, ErrTeachingForbidden
	}
	playlist.Title = strings.TrimSpace(req.Title)
	playlist.Description = strings.TrimSpace(req.Description)
	playlist.Visibility = req.Visibility
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&playlist).Error; err != nil {
			return err
		}
		return s.replacePlaylistProblems(tx, playlist.ID, req.Problems)
	}); err != nil {
		return nil, err
	}
	return s.GetPlaylistDetail(userID, role, playlist.ID, true)
}

func (s *TeachingService) DeletePlaylist(userID uint64, role string, id uint64) error {
	if !isTeacherRole(role) {
		return ErrTeachingForbidden
	}
	var playlist model.Playlist
	if err := s.db.First(&playlist, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPlaylistNotFound
		}
		return err
	}
	if role != "admin" && playlist.CreatedBy != userID {
		return ErrTeachingForbidden
	}
	var assignmentCount int64
	if err := s.db.Model(&model.Assignment{}).Where("playlist_id = ?", playlist.ID).Count(&assignmentCount).Error; err != nil {
		return err
	}
	if assignmentCount > 0 {
		return ErrPlaylistInUse
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("playlist_id = ?", playlist.ID).Delete(&model.PlaylistProblem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&playlist).Error
	})
}

func (s *TeachingService) ListMyClasses(userID uint64, role string) ([]dto.ClassroomListItem, error) {
	var items []dto.ClassroomListItem
	query := s.db.Table("classes c").
		Select("c.id, c.name, c.description, '' AS join_code, c.teacher_id, tu.username AS teacher_name, c.status, COALESCE(cm.role,'') AS member_role, COUNT(DISTINCT m.id) AS member_count, COUNT(DISTINCT a.id) AS assignment_count, c.created_at").
		Joins("JOIN users tu ON tu.id = c.teacher_id").
		Joins("LEFT JOIN class_members cm ON cm.class_id = c.id AND cm.user_id = ? AND cm.status = 'active'", userID).
		Joins("LEFT JOIN class_members m ON m.class_id = c.id AND m.status = 'active'").
		Joins("LEFT JOIN assignments a ON a.class_id = c.id")
	if role == "teacher" {
		query = query.Where("c.teacher_id = ? OR cm.user_id = ?", userID, userID)
	} else if role == "admin" {
		query = query.Where("1=1")
	} else {
		query = query.Where("cm.user_id = ?", userID)
	}
	if err := query.Group("c.id, tu.username, cm.role").Order("c.id DESC").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TeachingService) JoinClass(userID uint64, joinCode string) (*dto.ClassroomDetailResponse, error) {
	joinCode = strings.TrimSpace(strings.ToUpper(joinCode))
	var class model.Classroom
	if err := s.db.Where("join_code = ?", joinCode).First(&class).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJoinCodeInvalid
		}
		return nil, err
	}
	var existing model.ClassMember
	if err := s.db.Where("class_id = ? AND user_id = ?", class.ID, userID).First(&existing).Error; err == nil {
		if existing.Status != "removed" {
			return nil, ErrAlreadyJoinedClass
		}
		existing.Status = "active"
		existing.Role = "student"
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		member := model.ClassMember{ClassID: class.ID, UserID: userID, Role: "student", Status: "active"}
		if err := s.db.Create(&member).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return s.GetClassDetail(userID, "student", class.ID, false)
}

func (s *TeachingService) CreateClass(userID uint64, role string, req dto.CreateClassRequest) (*dto.ClassroomDetailResponse, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	class := model.Classroom{Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description), JoinCode: randomJoinCode(), TeacherID: userID, Status: "active"}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&class).Error; err != nil {
			return err
		}
		member := model.ClassMember{ClassID: class.ID, UserID: userID, Role: "assistant", Status: "active"}
		return tx.Create(&member).Error
	}); err != nil {
		return nil, err
	}
	return s.GetClassDetail(userID, role, class.ID, true)
}

func (s *TeachingService) UpdateClass(userID uint64, role string, id uint64, req dto.UpdateClassRequest) (*dto.ClassroomDetailResponse, error) {
	var class model.Classroom
	if err := s.db.First(&class, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}
	class.Name = strings.TrimSpace(req.Name)
	class.Description = strings.TrimSpace(req.Description)
	class.Status = req.Status
	if err := s.db.Save(&class).Error; err != nil {
		return nil, err
	}
	return s.GetClassDetail(userID, role, class.ID, true)
}

func (s *TeachingService) ListTeacherClasses(userID uint64, role string) ([]dto.ClassroomListItem, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	var items []dto.ClassroomListItem
	query := s.db.Table("classes c").
		Select("c.id, c.name, c.description, c.join_code, c.teacher_id, tu.username AS teacher_name, c.status, '' AS member_role, COUNT(DISTINCT m.id) AS member_count, COUNT(DISTINCT a.id) AS assignment_count, c.created_at").
		Joins("JOIN users tu ON tu.id = c.teacher_id").
		Joins("LEFT JOIN class_members m ON m.class_id = c.id AND m.status = 'active'").
		Joins("LEFT JOIN assignments a ON a.class_id = c.id")
	if role != "admin" {
		query = query.Where("c.teacher_id = ?", userID)
	}
	if err := query.Group("c.id, tu.username").Order("c.id DESC").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *TeachingService) GetClassDetail(userID uint64, role string, classID uint64, includeMembers bool) (*dto.ClassroomDetailResponse, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	canManage := s.canManageClass(userID, role, class)
	if !canManage {
		var member model.ClassMember
		if err := s.db.Where("class_id = ? AND user_id = ? AND status = 'active'", classID, userID).First(&member).Error; err != nil {
			return nil, ErrTeachingForbidden
		}
	}
	var teacher model.User
	if err := s.db.First(&teacher, class.TeacherID).Error; err != nil {
		return nil, err
	}
	assignments, err := s.listAssignmentsForClass(userID, classID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ClassroomDetailResponse{ID: class.ID, Name: class.Name, Description: class.Description, TeacherID: class.TeacherID, TeacherName: teacher.Username, Status: class.Status, CreatedAt: class.CreatedAt, UpdatedAt: class.UpdatedAt, Assignments: assignments}
	if canManage {
		resp.JoinCode = class.JoinCode
	}
	if includeMembers || canManage {
		members, err := s.listMembers(classID)
		if err != nil {
			return nil, err
		}
		resp.Members = members
	}
	return resp, nil
}

func (s *TeachingService) ListClassMembers(userID uint64, role string, classID uint64) ([]dto.ClassMemberItem, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}
	return s.listMembers(classID)
}

func (s *TeachingService) UpdateClassMember(userID uint64, role string, classID uint64, memberUserID uint64, req dto.UpdateClassMemberRequest) ([]dto.ClassMemberItem, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}
	var member model.ClassMember
	if err := s.db.Where("class_id = ? AND user_id = ?", classID, memberUserID).First(&member).Error; err != nil {
		return nil, err
	}
	member.Role = req.Role
	member.Status = req.Status
	if err := s.db.Save(&member).Error; err != nil {
		return nil, err
	}
	return s.listMembers(classID)
}

func (s *TeachingService) CreateAssignment(userID uint64, role string, classID uint64, req dto.CreateAssignmentRequest) (*dto.AssignmentDetailResponse, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}
	var playlist model.Playlist
	if err := s.db.First(&playlist, req.PlaylistID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	if role != "admin" && playlist.CreatedBy != userID && playlist.Visibility == "private" {
		return nil, ErrTeachingForbidden
	}
	assignment := model.Assignment{ClassID: classID, PlaylistID: req.PlaylistID, Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description), Type: req.Type, StartAt: req.StartAt, DueAt: req.DueAt, CreatedBy: userID}
	if err := s.db.Create(&assignment).Error; err != nil {
		return nil, err
	}
	return s.GetAssignmentDetail(userID, role, assignment.ID)
}

func (s *TeachingService) UpdateAssignment(userID uint64, role string, assignmentID uint64, req dto.UpdateAssignmentRequest) (*dto.AssignmentDetailResponse, error) {
	var assignment model.Assignment
	if err := s.db.First(&assignment, assignmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}

	var class model.Classroom
	if err := s.db.First(&class, assignment.ClassID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}

	var playlist model.Playlist
	if err := s.db.First(&playlist, req.PlaylistID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	if role != "admin" && playlist.CreatedBy != userID && playlist.Visibility == "private" {
		return nil, ErrTeachingForbidden
	}

	assignment.PlaylistID = req.PlaylistID
	assignment.Title = strings.TrimSpace(req.Title)
	assignment.Description = strings.TrimSpace(req.Description)
	assignment.Type = req.Type
	assignment.StartAt = req.StartAt
	assignment.DueAt = req.DueAt
	if err := s.db.Save(&assignment).Error; err != nil {
		return nil, err
	}
	return s.GetAssignmentDetail(userID, role, assignment.ID)
}

func (s *TeachingService) DeleteAssignment(userID uint64, role string, assignmentID uint64) error {
	var assignment model.Assignment
	if err := s.db.First(&assignment, assignmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAssignmentNotFound
		}
		return err
	}

	var class model.Classroom
	if err := s.db.First(&class, assignment.ClassID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrClassNotFound
		}
		return err
	}
	if !s.canManageClass(userID, role, class) {
		return ErrTeachingForbidden
	}
	return s.db.Delete(&assignment).Error
}

func (s *TeachingService) ListTeacherAssignments(userID uint64, role string, classID uint64) ([]dto.AssignmentListItem, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}
	return s.listAssignmentsForClass(userID, classID)
}

func (s *TeachingService) GetAssignmentDetail(userID uint64, role string, assignmentID uint64) (*dto.AssignmentDetailResponse, error) {
	var assignment model.Assignment
	if err := s.db.First(&assignment, assignmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}
	var class model.Classroom
	if err := s.db.First(&class, assignment.ClassID).Error; err != nil {
		return nil, err
	}
	canManage := s.canManageClass(userID, role, class)
	if !canManage {
		var member model.ClassMember
		if err := s.db.Where("class_id = ? AND user_id = ? AND status = 'active'", class.ID, userID).First(&member).Error; err != nil {
			return nil, ErrTeachingForbidden
		}
	}
	var playlist model.Playlist
	if err := s.db.First(&playlist, assignment.PlaylistID).Error; err != nil {
		return nil, err
	}
	var problems []dto.AssignmentProblemItem
	if err := s.db.Table("playlist_problems pp").Select("pp.problem_id, p.display_id, p.title, pp.display_order").Joins("JOIN problems p ON p.id = pp.problem_id").Where("pp.playlist_id = ?", assignment.PlaylistID).Order("pp.display_order ASC, pp.id ASC").Scan(&problems).Error; err != nil {
		return nil, err
	}
	statuses, err := s.loadAcceptedStatusMap(userID, problems)
	if err != nil {
		return nil, err
	}
	for i := range problems {
		if status, ok := statuses[problems[i].ProblemID]; ok {
			problems[i].Solved = status == "Accepted"
			problems[i].LastStatus = status
		}
	}
	solvedCount := 0
	for _, p := range problems {
		if p.Solved {
			solvedCount++
		}
	}
	return &dto.AssignmentDetailResponse{ID: assignment.ID, ClassID: assignment.ClassID, ClassName: class.Name, PlaylistID: assignment.PlaylistID, PlaylistTitle: playlist.Title, Title: assignment.Title, Description: assignment.Description, Type: assignment.Type, StartAt: assignment.StartAt, DueAt: assignment.DueAt, Status: assignmentStatus(assignment), SolvedCount: solvedCount, TotalCount: len(problems), Problems: problems}, nil
}

type classAnalyticsPlaylistProblemRow struct {
	PlaylistID uint64
	ProblemID  uint64
}

func (s *TeachingService) GetTeacherClassAnalytics(userID uint64, role string, classID uint64) (*dto.ClassAnalyticsResponse, error) {
	var class model.Classroom
	if err := s.db.First(&class, classID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}

	var teacher model.User
	if err := s.db.First(&teacher, class.TeacherID).Error; err != nil {
		return nil, err
	}

	var members []assignmentMemberRow
	if err := s.db.Table("class_members cm").
		Select("cm.user_id, u.username, u.userid AS user_id_no, cm.role").
		Joins("JOIN users u ON u.id = cm.user_id").
		Where("cm.class_id = ? AND cm.status = 'active'", class.ID).
		Order("cm.created_at ASC").
		Scan(&members).Error; err != nil {
		return nil, err
	}

	var assignments []model.Assignment
	if err := s.db.Where("class_id = ?", class.ID).Order("id DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}

	playlistIDs := make([]uint64, 0, len(assignments))
	for _, assignment := range assignments {
		playlistIDs = append(playlistIDs, assignment.PlaylistID)
	}

	playlistProblems := []classAnalyticsPlaylistProblemRow{}
	if len(playlistIDs) > 0 {
		if err := s.db.Table("playlist_problems").
			Select("playlist_id, problem_id").
			Where("playlist_id IN ?", playlistIDs).
			Scan(&playlistProblems).Error; err != nil {
			return nil, err
		}
	}

	playlistProblemMap := map[uint64][]uint64{}
	uniqueProblemSet := map[uint64]struct{}{}
	for _, row := range playlistProblems {
		playlistProblemMap[row.PlaylistID] = append(playlistProblemMap[row.PlaylistID], row.ProblemID)
		uniqueProblemSet[row.ProblemID] = struct{}{}
	}

	memberIDs := make([]uint64, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.UserID)
	}

	type classAnalyticsSubmissionRow struct {
		UserID    uint64
		ProblemID uint64
		Status    string
		CreatedAt time.Time
	}

	latestStatus := map[uint64]map[uint64]string{}
	acceptedMap := map[uint64]map[uint64]bool{}
	lastSubmissionAt := map[uint64]*time.Time{}

	if len(memberIDs) > 0 && len(uniqueProblemSet) > 0 {
		uniqueProblemIDs := make([]uint64, 0, len(uniqueProblemSet))
		for problemID := range uniqueProblemSet {
			uniqueProblemIDs = append(uniqueProblemIDs, problemID)
		}
		var rows []classAnalyticsSubmissionRow
		if err := s.db.Table("submissions").
			Select("user_id, problem_id, status, created_at").
			Where("user_id IN ? AND problem_id IN ? AND contest_id IS NULL", memberIDs, uniqueProblemIDs).
			Order("created_at DESC").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if lastSubmissionAt[row.UserID] == nil {
				t := row.CreatedAt
				lastSubmissionAt[row.UserID] = &t
			}
			if _, ok := latestStatus[row.UserID]; !ok {
				latestStatus[row.UserID] = map[uint64]string{}
			}
			if _, ok := acceptedMap[row.UserID]; !ok {
				acceptedMap[row.UserID] = map[uint64]bool{}
			}
			if _, exists := latestStatus[row.UserID][row.ProblemID]; !exists {
				latestStatus[row.UserID][row.ProblemID] = row.Status
			}
			if row.Status == "Accepted" {
				acceptedMap[row.UserID][row.ProblemID] = true
			}
		}
	}

	assignmentItems := make([]dto.ClassAnalyticsAssignmentItem, 0, len(assignments))
	topStudents := make([]dto.ClassAnalyticsTopStudentItem, 0, len(members))
	overallCompleted := 0
	completedAssignmentsByUser := map[uint64]int{}

	for _, assignment := range assignments {
		problemIDs := playlistProblemMap[assignment.PlaylistID]
		problemCount := len(problemIDs)
		startedCount := 0
		completedCount := 0
		for _, member := range members {
			started := false
			solvedCount := 0
			for _, problemID := range problemIDs {
				if latestStatus[member.UserID][problemID] != "" {
					started = true
				}
				if acceptedMap[member.UserID][problemID] {
					solvedCount++
				}
			}
			if started {
				startedCount++
			}
			if problemCount > 0 && solvedCount >= problemCount {
				completedCount++
				completedAssignmentsByUser[member.UserID]++
			}
		}
		if len(members) > 0 {
			overallCompleted += completedCount
		}
		completionRate := 0
		if len(members) > 0 {
			completionRate = int((float64(completedCount) / float64(len(members))) * 100)
		}
		assignmentItems = append(assignmentItems, dto.ClassAnalyticsAssignmentItem{
			AssignmentID:   assignment.ID,
			Title:          assignment.Title,
			Type:           assignment.Type,
			Status:         assignmentStatus(assignment),
			ProblemCount:   problemCount,
			StartedCount:   startedCount,
			CompletedCount: completedCount,
			CompletionRate: completionRate,
		})
	}

	for _, member := range members {
		solvedProblemCount := 0
		for problemID := range uniqueProblemSet {
			if acceptedMap[member.UserID][problemID] {
				solvedProblemCount++
			}
		}
		topStudents = append(topStudents, dto.ClassAnalyticsTopStudentItem{
			UserID:               member.UserID,
			Username:             member.Username,
			StudentID:            member.UserIDNo,
			Role:                 member.Role,
			SolvedProblemCount:   solvedProblemCount,
			CompletedAssignments: completedAssignmentsByUser[member.UserID],
			LastSubmissionAt:     lastSubmissionAt[member.UserID],
		})
	}

	sort.Slice(topStudents, func(i, j int) bool {
		if topStudents[i].CompletedAssignments != topStudents[j].CompletedAssignments {
			return topStudents[i].CompletedAssignments > topStudents[j].CompletedAssignments
		}
		if topStudents[i].SolvedProblemCount != topStudents[j].SolvedProblemCount {
			return topStudents[i].SolvedProblemCount > topStudents[j].SolvedProblemCount
		}
		if topStudents[i].LastSubmissionAt == nil && topStudents[j].LastSubmissionAt != nil {
			return false
		}
		if topStudents[i].LastSubmissionAt != nil && topStudents[j].LastSubmissionAt == nil {
			return true
		}
		if topStudents[i].LastSubmissionAt != nil && topStudents[j].LastSubmissionAt != nil && !topStudents[i].LastSubmissionAt.Equal(*topStudents[j].LastSubmissionAt) {
			return topStudents[i].LastSubmissionAt.After(*topStudents[j].LastSubmissionAt)
		}
		return strings.ToLower(topStudents[i].Username) < strings.ToLower(topStudents[j].Username)
	})
	if len(topStudents) > 8 {
		topStudents = topStudents[:8]
	}

	overallCompletionRate := 0
	if len(assignments) > 0 && len(members) > 0 {
		overallCompletionRate = int((float64(overallCompleted) / float64(len(assignments)*len(members))) * 100)
	}

	return &dto.ClassAnalyticsResponse{
		ClassID:               class.ID,
		ClassName:             class.Name,
		TeacherName:           teacher.Username,
		MemberCount:           len(members),
		AssignmentCount:       len(assignments),
		UniqueProblemCount:    len(uniqueProblemSet),
		OverallCompletionRate: overallCompletionRate,
		Assignments:           assignmentItems,
		TopStudents:           topStudents,
	}, nil
}

func (s *TeachingService) validatePlaylistProblems(problems []dto.PlaylistProblemRequest) error {
	ids := make([]uint64, 0, len(problems))
	for _, item := range problems {
		ids = append(ids, item.ProblemID)
	}
	var count int64
	if err := s.db.Model(&model.Problem{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return ErrProblemNotFound
	}
	return nil
}

func (s *TeachingService) replacePlaylistProblems(tx *gorm.DB, playlistID uint64, problems []dto.PlaylistProblemRequest) error {
	if err := tx.Where("playlist_id = ?", playlistID).Delete(&model.PlaylistProblem{}).Error; err != nil {
		return err
	}
	rows := make([]model.PlaylistProblem, 0, len(problems))
	for _, item := range problems {
		rows = append(rows, model.PlaylistProblem{PlaylistID: playlistID, ProblemID: item.ProblemID, DisplayOrder: item.DisplayOrder})
	}
	return tx.Create(&rows).Error
}

func (s *TeachingService) listMembers(classID uint64) ([]dto.ClassMemberItem, error) {
	var members []dto.ClassMemberItem
	err := s.db.Table("class_members cm").Select("cm.user_id, u.username, u.userid, cm.role, cm.status, cm.created_at AS joined_at").Joins("JOIN users u ON u.id = cm.user_id").Where("cm.class_id = ?", classID).Order("cm.created_at ASC").Scan(&members).Error
	return members, err
}

func (s *TeachingService) listAssignmentsForClass(userID uint64, classID uint64) ([]dto.AssignmentListItem, error) {
	var assignments []model.Assignment
	if err := s.db.Where("class_id = ?", classID).Order("id DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}
	items := make([]dto.AssignmentListItem, 0, len(assignments))
	for _, assignment := range assignments {
		var total int64
		if err := s.db.Model(&model.PlaylistProblem{}).Where("playlist_id = ?", assignment.PlaylistID).Count(&total).Error; err != nil {
			return nil, err
		}
		var solved int64
		err := s.db.Table("submissions s").Select("COUNT(DISTINCT s.problem_id)").Joins("JOIN playlist_problems pp ON pp.problem_id = s.problem_id").Where("pp.playlist_id = ? AND s.user_id = ? AND s.status = ?", assignment.PlaylistID, userID, "Accepted").Scan(&solved).Error
		if err != nil {
			return nil, err
		}
		items = append(items, dto.AssignmentListItem{ID: assignment.ID, ClassID: assignment.ClassID, PlaylistID: assignment.PlaylistID, Title: assignment.Title, Description: assignment.Description, Type: assignment.Type, StartAt: assignment.StartAt, DueAt: assignment.DueAt, Status: assignmentStatus(assignment), SolvedCount: int(solved), TotalCount: int(total), CreatedAt: assignment.CreatedAt})
	}
	return items, nil
}

func (s *TeachingService) loadAcceptedStatusMap(userID uint64, problems []dto.AssignmentProblemItem) (map[uint64]string, error) {
	statusMap := map[uint64]string{}
	if userID == 0 || len(problems) == 0 {
		return statusMap, nil
	}
	ids := make([]uint64, 0, len(problems))
	for _, item := range problems {
		ids = append(ids, item.ProblemID)
	}
	var rows []assignmentStatusRow
	if err := s.db.Table("submissions").Select("problem_id, status, created_at").Where("user_id = ? AND problem_id IN ?", userID, ids).Order("created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if _, exists := statusMap[row.ProblemID]; exists {
			continue
		}
		statusMap[row.ProblemID] = row.Status
	}
	return statusMap, nil
}

func (s *TeachingService) canManageClass(userID uint64, role string, class model.Classroom) bool {
	return role == "admin" || (role == "teacher" && class.TeacherID == userID)
}

func isTeacherRole(role string) bool {
	return role == "teacher" || role == "admin"
}

func assignmentStatus(a model.Assignment) string {
	now := time.Now()
	if a.StartAt != nil && now.Before(*a.StartAt) {
		return "upcoming"
	}
	if a.DueAt != nil && now.After(*a.DueAt) {
		return "closed"
	}
	return "open"
}

func randomJoinCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = charset[r.Intn(len(charset))]
	}
	return string(buf)
}

func (s *TeachingService) EnsureUniqueJoinCode() string {
	for {
		code := randomJoinCode()
		var count int64
		_ = s.db.Model(&model.Classroom{}).Where("join_code = ?", code).Count(&count).Error
		if count == 0 {
			return code
		}
	}
}

func (s *TeachingService) CreateClassWithUniqueCode(userID uint64, role string, req dto.CreateClassRequest) (*dto.ClassroomDetailResponse, error) {
	if !isTeacherRole(role) {
		return nil, ErrTeachingForbidden
	}
	class := model.Classroom{Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description), JoinCode: s.EnsureUniqueJoinCode(), TeacherID: userID, Status: "active"}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&class).Error; err != nil {
			return err
		}
		member := model.ClassMember{ClassID: class.ID, UserID: userID, Role: "assistant", Status: "active"}
		return tx.Create(&member).Error
	}); err != nil {
		return nil, err
	}
	return s.GetClassDetail(userID, role, class.ID, true)
}

type assignmentMemberRow struct {
	UserID   uint64
	Username string
	UserIDNo string
	Role     string
}

type assignmentSubmissionRow struct {
	UserID    uint64
	ProblemID uint64
	Status    string
	CreatedAt time.Time
}

func (s *TeachingService) GetTeacherAssignmentOverview(userID uint64, role string, assignmentID uint64) (*dto.AssignmentOverviewResponse, error) {
	var assignment model.Assignment
	if err := s.db.First(&assignment, assignmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}

	var class model.Classroom
	if err := s.db.First(&class, assignment.ClassID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if !s.canManageClass(userID, role, class) {
		return nil, ErrTeachingForbidden
	}

	var playlist model.Playlist
	if err := s.db.First(&playlist, assignment.PlaylistID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}

	var problems []dto.AssignmentProblemItem
	if err := s.db.Table("playlist_problems pp").
		Select("pp.problem_id, p.display_id, p.title, pp.display_order").
		Joins("JOIN problems p ON p.id = pp.problem_id").
		Where("pp.playlist_id = ?", assignment.PlaylistID).
		Order("pp.display_order ASC, pp.id ASC").
		Scan(&problems).Error; err != nil {
		return nil, err
	}

	var members []assignmentMemberRow
	if err := s.db.Table("class_members cm").
		Select("cm.user_id, u.username, u.userid AS user_id_no, cm.role").
		Joins("JOIN users u ON u.id = cm.user_id").
		Where("cm.class_id = ? AND cm.status = 'active'", class.ID).
		Order("cm.created_at ASC").
		Scan(&members).Error; err != nil {
		return nil, err
	}

	problemIDs := make([]uint64, 0, len(problems))
	for _, item := range problems {
		problemIDs = append(problemIDs, item.ProblemID)
	}
	memberIDs := make([]uint64, 0, len(members))
	for _, item := range members {
		memberIDs = append(memberIDs, item.UserID)
	}

	latestStatus := map[uint64]map[uint64]string{}
	acceptedMap := map[uint64]map[uint64]bool{}
	lastSubmissionAt := map[uint64]*time.Time{}

	if len(problemIDs) > 0 && len(memberIDs) > 0 {
		var rows []assignmentSubmissionRow
		if err := s.db.Table("submissions").
			Select("user_id, problem_id, status, created_at").
			Where("user_id IN ? AND problem_id IN ? AND contest_id IS NULL", memberIDs, problemIDs).
			Order("created_at DESC").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if lastSubmissionAt[row.UserID] == nil {
				t := row.CreatedAt
				lastSubmissionAt[row.UserID] = &t
			}
			if _, ok := latestStatus[row.UserID]; !ok {
				latestStatus[row.UserID] = map[uint64]string{}
			}
			if _, ok := acceptedMap[row.UserID]; !ok {
				acceptedMap[row.UserID] = map[uint64]bool{}
			}
			if _, exists := latestStatus[row.UserID][row.ProblemID]; !exists {
				latestStatus[row.UserID][row.ProblemID] = row.Status
			}
			if row.Status == "Accepted" {
				acceptedMap[row.UserID][row.ProblemID] = true
			}
		}
	}

	memberItems := make([]dto.AssignmentStudentProgressItem, 0, len(members))
	completedCount := 0
	startedCount := 0
	problemCount := len(problems)

	for _, member := range members {
		problemStatuses := make([]dto.AssignmentStudentProblemStatus, 0, len(problems))
		solvedCount := 0
		started := false
		for _, problem := range problems {
			solved := acceptedMap[member.UserID][problem.ProblemID]
			status := latestStatus[member.UserID][problem.ProblemID]
			if solved {
				solvedCount++
			}
			if status != "" {
				started = true
			}
			problemStatuses = append(problemStatuses, dto.AssignmentStudentProblemStatus{
				ProblemID:    problem.ProblemID,
				DisplayID:    problem.DisplayID,
				Title:        problem.Title,
				DisplayOrder: problem.DisplayOrder,
				Solved:       solved,
				LastStatus:   status,
			})
		}
		progressStatus := "not_started"
		if problemCount > 0 && solvedCount >= problemCount {
			progressStatus = "completed"
			completedCount++
		} else if started || solvedCount > 0 {
			progressStatus = "in_progress"
			startedCount++
		}
		memberItems = append(memberItems, dto.AssignmentStudentProgressItem{
			UserID:           member.UserID,
			Username:         member.Username,
			StudentID:        member.UserIDNo,
			Role:             member.Role,
			SolvedCount:      solvedCount,
			TotalCount:       problemCount,
			ProgressStatus:   progressStatus,
			LastSubmissionAt: lastSubmissionAt[member.UserID],
			ProblemStatuses:  problemStatuses,
		})
	}

	return &dto.AssignmentOverviewResponse{
		ID:             assignment.ID,
		ClassID:        assignment.ClassID,
		ClassName:      class.Name,
		PlaylistID:     assignment.PlaylistID,
		PlaylistTitle:  playlist.Title,
		Title:          assignment.Title,
		Description:    assignment.Description,
		Type:           assignment.Type,
		StartAt:        assignment.StartAt,
		DueAt:          assignment.DueAt,
		Status:         assignmentStatus(assignment),
		MemberCount:    len(memberItems),
		CompletedCount: completedCount,
		StartedCount:   startedCount,
		ProblemCount:   problemCount,
		Members:        memberItems,
	}, nil
}
