package services

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type UserListRequest struct {
	Keyword    string `form:"keyword"`
	Type       string `form:"type"`
	Role       string `form:"role"`
	Status     string `form:"status"`
	SyncStatus string `form:"syncStatus"`
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
}

type UserDTO struct {
	ID               string   `json:"id"`
	Username         string   `json:"username"`
	RealName         string   `json:"realName"`
	Gender           string   `json:"gender"`
	Type             string   `json:"type"`
	AccountType      string   `json:"accountType"`
	Phone            string   `json:"phone"`
	Email            string   `json:"email"`
	Birthdate        string   `json:"birthdate"`
	Status           string   `json:"status"`
	Sort             int      `json:"sort"`
	IdNumber         string   `json:"idNumber"`
	ICNumber         string   `json:"icNumber"`
	Avatar           string   `json:"avatar"`
	IsCreateAccount  bool     `json:"isCreateAccount"`
	BindStatus       string   `json:"bindStatus"`
	IsSyncCloud      bool     `json:"isSyncCloud"`
	SyncStatus       string   `json:"syncStatus"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	Role             string   `json:"role"`
	Roles            []string `json:"roles"`
	RoleNames        []string `json:"roleNames"`
	HasSignature     bool     `json:"hasSignature"`
	SignatureImageID string  `json:"signatureImageId,omitempty"`
	SignatureImage   string  `json:"signatureImage,omitempty"`
	DepartmentID     *int64  `json:"departmentId"`
}

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: database.GetDB()}
}

type rawUserRow struct {
	ID              int64  `gorm:"column:Id"`
	UserName        string `gorm:"column:UserName"`
	Name            string `gorm:"column:Name"`
	Gender          string `gorm:"column:Gender"`
	Type            string `gorm:"column:Type"`
	Phone           string `gorm:"column:Phone"`
	Email           string `gorm:"column:Email"`
	Birthdate       string `gorm:"column:Birthdate"`
	Sort            int    `gorm:"column:Sort"`
	IsDisabled      bool   `gorm:"column:IsDisabled"`
	IsDeleted       bool   `gorm:"column:IsDeleted"`
	IsCreateAccount bool   `gorm:"column:IsCreateAccount"`
	IsSyncCloud     bool   `gorm:"column:IsSyncCloud"`
	IdNumber        string `gorm:"column:IdNumber"`
	ICNumber        string `gorm:"column:ICNumber"`
	Avatar          string `gorm:"column:Avatar"`
	CreationTime    string `gorm:"column:CreationTime"`
	LastModifyTime  string `gorm:"column:LastModifyTime"`
	RoleName        string `gorm:"column:RoleName"`
	SigID           int64  `gorm:"column:SigID"`
	SigLen          int64  `gorm:"column:SigLen"`
	TotalCount      int64  `gorm:"column:TotalCount"`
}

func (s *UserService) buildUserBaseQuery() *gorm.DB {
	return s.db.Table(`"Identity_Users" AS u`).
		Select(`u."Id", u."UserName",
			COALESCE(e."Name", '') AS "Name",
			COALESCE(e."Gender", '') AS "Gender",
			COALESCE(e."Type", '') AS "Type",
			COALESCE(NULLIF(e."PhoneNumber", ''), u."PhoneNumber") AS "Phone",
			COALESCE(NULLIF(e."Email", ''), u."Email") AS "Email",
			COALESCE(e."Birthdate"::text, '') AS "Birthdate",
			COALESCE(e."Sort", 0) AS "Sort",
			COALESCE(e."IsDisabled", false) AS "IsDisabled",
			COALESCE(e."IsDeleted", false) AS "IsDeleted",
			COALESCE(e."IsCreateAccount", false) AS "IsCreateAccount",
			COALESCE(e."IsSyncCloud", false) AS "IsSyncCloud",
			COALESCE(e."IdNumber", '') AS "IdNumber",
			COALESCE(e."ICNumber", '') AS "ICNumber",
			COALESCE(e."Avatar", '') AS "Avatar",
			COALESCE(e."CreationTime"::text, '') AS "CreationTime",
			COALESCE(e."LastModifyTime"::text, e."LastModificationTime"::text, '') AS "LastModifyTime",
			COALESCE(r."Name", '') AS "RoleName",
			COALESCE(sig."Id", 0) AS "SigID",
			COALESCE(length(sig."ImageBase64String"), 0) AS "SigLen"`).
		Joins(`LEFT JOIN "Organ_Employee" AS e ON e."Id" = u."Id"`).
		Joins(`LEFT JOIN "Identity_UserRoles" AS ur ON ur."UserId" = u."Id"`).
		Joins(`LEFT JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Joins(`LEFT JOIN LATERAL (
			SELECT ui."Id", ui."ImageBase64String"
			FROM "User_Image" ui
			WHERE ui."UserId" = u."Id" AND ui."Type" = 10
			ORDER BY ui."LastModifyTime" DESC, ui."Id" DESC
			LIMIT 1
		) AS sig ON true`)
}

func (s *UserService) buildUserQuery() *gorm.DB {
	return s.buildUserBaseQuery().Where(`COALESCE(e."IsDeleted", false) = false`)
}

func (s *UserService) List(req UserListRequest) (*UserListResponse, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	type countRow struct {
		Total int64 `gorm:"column:TotalCount"`
	}

	baseQuery := s.buildUserQuery()

	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		baseQuery = baseQuery.Where(
			`(COALESCE(e."Name", '') ILIKE ? OR u."UserName" ILIKE ? OR COALESCE(e."PhoneNumber", u."PhoneNumber") ILIKE ? OR COALESCE(e."IdNumber", '') ILIKE ?)`,
			kw, kw, kw, kw)
	}
	if req.Type != "" {
		baseQuery = baseQuery.Where(`COALESCE(e."Type", '') = ?`, req.Type)
	}
	if req.Role != "" {
		baseQuery = baseQuery.Where(`r."Name" = ?`, req.Role)
	}
	if req.Status == "active" {
		baseQuery = baseQuery.Where(`COALESCE(e."IsDisabled", false) = false`)
	} else if req.Status == "disabled" {
		baseQuery = baseQuery.Where(`COALESCE(e."IsDisabled", false) = true`)
	}
	if req.SyncStatus == "synced" {
		baseQuery = baseQuery.Where(`COALESCE(e."IsSyncCloud", false) = true`)
	} else if req.SyncStatus == "unsynced" {
		baseQuery = baseQuery.Where(`COALESCE(e."IsSyncCloud", false) = false`)
	}

	var cnt countRow
	if err := s.db.Table("(?) AS sub", baseQuery).Select(`COUNT(DISTINCT "Id") AS "TotalCount"`).Find(&cnt).Error; err != nil {
		return nil, fmt.Errorf("查询用户总数失败: %w", err)
	}
	total := cnt.Total

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	orderClause := fmt.Sprintf(`"Id" ASC LIMIT %d OFFSET %d`, pageSize, offset)

	rows := []rawUserRow{}
	err := baseQuery.Order(orderClause).Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	seen := make(map[int64]*UserDTO)
	order := make([]int64, 0, len(rows))
	for _, row := range rows {
		if dto, ok := seen[row.ID]; ok {
			if dto.Role == "" && row.RoleName != "" {
				dto.Role = row.RoleName
			}
			if row.RoleName != "" {
				dto.RoleNames = appendIfNotExists(dto.RoleNames, row.RoleName)
			}
			continue
		}
		birthdate := ""
		if row.Birthdate != "" {
			birthdate = row.Birthdate[:min(10, len(row.Birthdate))]
		}
		createdAt := ""
		if row.CreationTime != "" {
			createdAt = row.CreationTime[:min(19, len(row.CreationTime))]
		}
		updatedAt := ""
		if row.LastModifyTime != "" {
			updatedAt = row.LastModifyTime[:min(19, len(row.LastModifyTime))]
		}
		status := "active"
		if row.IsDisabled {
			status = "disabled"
		}
		bindStatus := "unbound"
		if row.IsCreateAccount {
			bindStatus = "bound"
		}
		syncStatus := "unsynced"
		if row.IsSyncCloud {
			syncStatus = "synced"
		}
		roleNames := []string{}
		if row.RoleName != "" {
			roleNames = append(roleNames, row.RoleName)
		}
		dto := &UserDTO{
			ID:               fmt.Sprintf("%d", row.ID),
			Username:         row.UserName,
			RealName:         row.Name,
			Gender:           row.Gender,
			Type:             row.Type,
			AccountType:      "",
			Phone:            row.Phone,
			Email:            row.Email,
			Birthdate:        birthdate,
			Status:           status,
			Sort:             row.Sort,
			IdNumber:         row.IdNumber,
			ICNumber:         row.ICNumber,
			Avatar:           row.Avatar,
			IsCreateAccount:  row.IsCreateAccount,
			BindStatus:       bindStatus,
			IsSyncCloud:      row.IsSyncCloud,
			SyncStatus:       syncStatus,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			Role:             row.RoleName,
			RoleNames:        roleNames,
			HasSignature:     row.SigID > 0 && row.SigLen > 0,
			SignatureImageID: fmt.Sprintf("%d", row.SigID),
		}
		seen[row.ID] = dto
		order = append(order, row.ID)
	}

	result := make([]UserDTO, 0, len(order))
	for _, id := range order {
		dto := seen[id]
		if req.Role != "" && dto.Role != req.Role {
			found := false
			for _, rn := range dto.RoleNames {
				if rn == req.Role {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		result = append(result, *dto)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &UserListResponse{
		Items:      result,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

type UserListResponse struct {
	Items      []UserDTO `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	TotalPages int       `json:"totalPages"`
}

func appendIfNotExists(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *UserService) GetByID(id string) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	rows := []rawUserRow{}
	err := s.buildUserQuery().Where(`u."Id" = ?`, id).Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	row := rows[0]
	birthdate := ""
	if row.Birthdate != "" {
		birthdate = row.Birthdate[:min(10, len(row.Birthdate))]
	}
	createdAt := ""
	if row.CreationTime != "" {
		createdAt = row.CreationTime[:min(19, len(row.CreationTime))]
	}
	updatedAt := ""
	if row.LastModifyTime != "" {
		updatedAt = row.LastModifyTime[:min(19, len(row.LastModifyTime))]
	}
	status := "active"
	if row.IsDisabled {
		status = "disabled"
	}
	bindStatus := "unbound"
	if row.IsCreateAccount {
		bindStatus = "bound"
	}
	syncStatus := "unsynced"
	if row.IsSyncCloud {
		syncStatus = "synced"
	}
	roleNames := []string{}
	if row.RoleName != "" {
		roleNames = append(roleNames, row.RoleName)
	}

	dto := &UserDTO{
		ID:               fmt.Sprintf("%d", row.ID),
		Username:         row.UserName,
		RealName:         row.Name,
		Gender:           row.Gender,
		Type:             row.Type,
		AccountType:      "",
		Phone:            row.Phone,
		Email:            row.Email,
		Birthdate:        birthdate,
		Status:           status,
		Sort:             row.Sort,
		IdNumber:         row.IdNumber,
		ICNumber:         row.ICNumber,
		Avatar:           row.Avatar,
		IsCreateAccount:  row.IsCreateAccount,
		BindStatus:       bindStatus,
		IsSyncCloud:      row.IsSyncCloud,
		SyncStatus:       syncStatus,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		Role:             row.RoleName,
		RoleNames:        roleNames,
		HasSignature:     row.SigID > 0 && row.SigLen > 0,
		SignatureImageID: fmt.Sprintf("%d", row.SigID),
	}

	if row.SigID > 0 && row.SigLen > 0 {
		var sig struct {
			ImageBase64String string `gorm:"column:ImageBase64String"`
		}
		s.db.Table(`"User_Image"`).Where(`"Id" = ?`, row.SigID).Select(`"ImageBase64String"`).First(&sig)
		dto.SignatureImage = sig.ImageBase64String
	}

	for _, r := range rows {
		if r.RoleName != "" && r.RoleName != dto.Role {
			dto.RoleNames = appendIfNotExists(dto.RoleNames, r.RoleName)
		}
	}

	return dto, nil
}

func (s *UserService) Create(req UserCreateRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var maxID struct{ Id int64 }
	tx.Table(`"Identity_Users"`).Select(`COALESCE(MAX("Id"), 300000)`).Scan(&maxID)
	newID := maxID.Id + 1

	now := time.Now()
	securityStamp := randomUUID()
	concurrencyStamp := randomUUID()

	idColumns := map[string]interface{}{
		`"Id"`:                   newID,
		`"UserName"`:             req.Username,
		`"NormalizedUserName"`:   strings.ToUpper(req.Username),
		`"Email"`:                req.Email,
		`"NormalizedEmail"`:      strings.ToUpper(req.Email),
		`"EmailConfirmed"`:       false,
		`"PhoneNumber"`:          req.PhoneNumber,
		`"PhoneNumberConfirmed"`: false,
		`"TwoFactorEnabled"`:     false,
		`"LockoutEnabled"`:       true,
		`"AccessFailedCount"`:    0,
		`"SecurityStamp"`:        securityStamp,
		`"ConcurrencyStamp"`:     concurrencyStamp,
		`"LastModifyTime"`:       now,
	}
	if req.Password != "" {
		hash, err := HashASPNetIdentityV3Password(req.Password)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("生成密码哈希失败: %w", err)
		}
		idColumns[`"PasswordHash"`] = hash
	}

	if err := tx.Table(`"Identity_Users"`).Create(idColumns).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	employeeName := req.RealName
	if employeeName == "" {
		employeeName = req.Username
	}

	empColumns := map[string]interface{}{
		`"Id"`:                   newID,
		`"Name"`:                 employeeName,
		`"Gender"`:               req.Gender,
		`"Type"`:                 req.Type,
		`"Birthdate"`:            timeToNullString(req.Birthdate),
		`"Avatar"`:               req.Avatar,
		`"Sort"`:                 req.Sort,
		`"IsDisabled"`:           false,
		`"IsDeleted"`:            false,
		`"CreationTime"`:         now,
		`"CreatorId"`:            req.CreatorID,
		`"LastModificationTime"`: now,
		`"LastModifierId"`:       req.CreatorID,
		`"PhoneNumber"`:          req.PhoneNumber,
		`"Email"`:                req.Email,
		`"IsCreateAccount"`:      true,
		`"IdNumber"`:             req.IdNumber,
		`"ICNumber"`:             req.ICNumber,
		`"LastModifyTime"`:       now,
		`"IsSyncCloud"`:          false,
	}

	if err := tx.Table(`"Organ_Employee"`).Create(empColumns).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建人员信息失败: %w", err)
	}

	if len(req.Roles) > 0 {
		for _, roleID := range req.Roles {
			roleCol := map[string]interface{}{
				`"UserId"`: newID,
				`"RoleId"`: roleID,
			}
			if err := tx.Table(`"Identity_UserRoles"`).Create(roleCol).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("绑定角色失败: %w", err)
			}
		}
	}

	if req.SignatureImage != "" {
		sigCol := map[string]interface{}{
			`"TenantId"`:          LegacyTenantID,
			`"UserId"`:            newID,
			`"ImageBase64String"`: req.SignatureImage,
			`"CreatorId"`:         req.CreatorID,
			`"CreateTime"`:        now,
			`"LastModifyTime"`:    now,
			`"Type"`:              10,
		}
		if err := tx.Table(`"User_Image"`).Create(sigCol).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("保存签名失败: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return s.GetByID(fmt.Sprintf("%d", newID))
}

func timeToNullString(t string) interface{} {
	if t == "" {
		return nil
	}
	return t
}

func (s *UserService) Update(id string, req UserUpdateRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	idColumns := map[string]interface{}{}
	if req.Username != "" {
		idColumns[`"UserName"`] = req.Username
		idColumns[`"NormalizedUserName"`] = strings.ToUpper(req.Username)
	}
	if req.Email != "" {
		idColumns[`"Email"`] = req.Email
		idColumns[`"NormalizedEmail"`] = strings.ToUpper(req.Email)
	}
	if req.PhoneNumber != "" {
		idColumns[`"PhoneNumber"`] = req.PhoneNumber
	}
	if len(idColumns) > 0 {
		idColumns[`"LastModifyTime"`] = now
		if err := tx.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Updates(idColumns).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新用户失败: %w", err)
		}
	}

	empColumns := map[string]interface{}{}
	if req.RealName != "" {
		empColumns[`"Name"`] = req.RealName
	}
	if req.Gender != "" {
		empColumns[`"Gender"`] = req.Gender
	}
	if req.Type != "" {
		empColumns[`"Type"`] = req.Type
	}
	if req.PhoneNumber != "" {
		empColumns[`"PhoneNumber"`] = req.PhoneNumber
	}
	if req.Email != "" {
		empColumns[`"Email"`] = req.Email
	}
	if req.Sort != nil {
		empColumns[`"Sort"`] = *req.Sort
	}
	if req.IdNumber != "" {
		empColumns[`"IdNumber"`] = req.IdNumber
	}
	if req.ICNumber != "" {
		empColumns[`"ICNumber"`] = req.ICNumber
	}
	if req.Avatar != "" {
		empColumns[`"Avatar"`] = req.Avatar
	}
	if req.Birthdate != "" {
		empColumns[`"Birthdate"`] = req.Birthdate
	}
	if len(empColumns) > 0 {
		empColumns[`"LastModificationTime"`] = now
		empColumns[`"LastModifierId"`] = req.CreatorID
		empColumns[`"LastModifyTime"`] = now
		if err := tx.Table(`"Organ_Employee"`).Where(`"Id" = ?`, id).Updates(empColumns).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新人员信息失败: %w", err)
		}
	}

	if len(req.Roles) > 0 {
		tx.Table(`"Identity_UserRoles"`).Where(`"UserId" = ?`, id).Delete(nil)
		for _, roleID := range req.Roles {
			roleCol := map[string]interface{}{
				`"UserId"`: id,
				`"RoleId"`: roleID,
			}
			if err := tx.Table(`"Identity_UserRoles"`).Create(roleCol).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("绑定角色失败: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return s.GetByID(id)
}

func (s *UserService) UpdateStatus(id string, req UserUpdateStatusRequest) (*UserDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	isDisabled := req.Status == "disabled"
	now := time.Now()
	columns := map[string]interface{}{
		`"IsDisabled"`:           isDisabled,
		`"LastModifyTime"`:       now,
		`"LastModificationTime"`: now,
	}
	result := s.db.Table(`"Organ_Employee"`).Where(`"Id" = ? AND COALESCE("IsDeleted", false) = false`, id).Updates(columns)
	if result.Error != nil {
		return nil, fmt.Errorf("更新用户状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return s.GetByID(id)
}

func (s *UserService) Delete(id string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	now := time.Now()
	columns := map[string]interface{}{
		`"IsDeleted"`:            true,
		`"IsDisabled"`:           true,
		`"LastModifyTime"`:       now,
		`"LastModificationTime"`: now,
	}
	result := s.db.Table(`"Organ_Employee"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return fmt.Errorf("删除用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *UserService) ResetPassword(id string, newPassword string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	hashedPassword, err := HashASPNetIdentityV3Password(newPassword)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}
	columns := map[string]interface{}{
		`"PasswordHash"`:       hashedPassword,
		`"AccessFailedCount"`:  0,
		`"LockoutEnd"`:         nil,
	}
	result := s.db.Table(`"Identity_Users"`).Where(`"Id" = ?`, id).Updates(columns)
	if result.Error != nil {
		return fmt.Errorf("重置密码失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *UserService) GetUserRoles(userID string) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	type roleRow struct {
		RoleName string `gorm:"column:RoleName"`
	}
	var rows []roleRow
	err := s.db.Table(`"Identity_UserRoles" AS ur`).
		Select(`r."Name" AS "RoleName"`).
		Joins(`JOIN "Identity_Roles" AS r ON r."Id" = ur."RoleId"`).
		Where(`ur."UserId" = ?`, userID).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}
	roles := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.RoleName != "" {
			roles = append(roles, row.RoleName)
		}
	}
	return roles, nil
}

func (s *UserService) SetUserRoles(userID string, roleIDs []string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	s.db.Table(`"Identity_UserRoles"`).Where(`"UserId" = ?`, userID).Delete(nil)
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		columns := map[string]interface{}{
			`"UserId"`: userID,
			`"RoleId"`: roleID,
		}
		s.db.Table(`"Identity_UserRoles"`).Create(columns)
	}
	return nil
}

func (s *UserService) GetSignature(userID string) (*UserSignatureDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var sig struct {
		ID                int64  `gorm:"column:Id"`
		ImageBase64String string `gorm:"column:ImageBase64String"`
	}
	err := s.db.Table(`"User_Image"`).
		Select(`"Id", "ImageBase64String"`).
		Where(`"UserId" = ? AND "Type" = 10`, userID).
		Order(`"LastModifyTime" DESC, "Id" DESC`).
		Limit(1).
		First(&sig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &UserSignatureDTO{UserID: userID}, nil
		}
		return nil, fmt.Errorf("查询签名失败: %w", err)
	}
	return &UserSignatureDTO{
		UserID:         userID,
		SignatureID:    fmt.Sprintf("%d", sig.ID),
		SignatureImage: sig.ImageBase64String,
	}, nil
}

type UserSignatureDTO struct {
	UserID         string `json:"userId"`
	SignatureID    string `json:"signatureId,omitempty"`
	SignatureImage string `json:"signatureImage,omitempty"`
}

func (s *UserService) UpdateSignature(userID string, imageBase64 string) (*UserSignatureDTO, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	now := time.Now()
	var existing struct{ Id int64 }
	err := s.db.Table(`"User_Image"`).
		Select(`"Id"`).
		Where(`"UserId" = ? AND "Type" = 10`, userID).
		Order(`"LastModifyTime" DESC, "Id" DESC`).
		Limit(1).
		First(&existing).Error

	if err == nil {
		upd := map[string]interface{}{
			`"ImageBase64String"`: imageBase64,
			`"LastModifyTime"`:    now,
		}
		if uerr := s.db.Table(`"User_Image"`).Where(`"Id" = ?`, existing.Id).Updates(upd).Error; uerr != nil {
			return nil, fmt.Errorf("更新签名失败: %w", uerr)
		}
		return &UserSignatureDTO{
			UserID:         userID,
			SignatureID:    fmt.Sprintf("%d", existing.Id),
			SignatureImage: imageBase64,
		}, nil
	}

	insert := map[string]interface{}{
		`"TenantId"`:          LegacyTenantID,
		`"UserId"`:            userID,
		`"ImageBase64String"`: imageBase64,
		`"CreatorId"`:         0,
		`"CreateTime"`:        now,
		`"LastModifyTime"`:    now,
		`"Type"`:              10,
	}
	result := s.db.Table(`"User_Image"`).Create(insert)
	if result.Error != nil {
		return nil, fmt.Errorf("保存签名失败: %w", result.Error)
	}
	var newID int64
	s.db.Raw(`SELECT LASTVAL()`).Scan(&newID)
	if newID == 0 {
		var maxID int64
		s.db.Table(`"User_Image"`).Select(`MAX("Id")`).Scan(&maxID)
		newID = maxID
	}
	return &UserSignatureDTO{
		UserID:         userID,
		SignatureID:    fmt.Sprintf("%d", newID),
		SignatureImage: imageBase64,
	}, nil
}

func (s *UserService) DeleteSignature(userID string) error {
	if s.db == nil {
		return errors.New("database not available")
	}
	result := s.db.Table(`"User_Image"`).Where(`"UserId" = ? AND "Type" = 10`, userID).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除签名失败: %w", result.Error)
	}
	return nil
}

func (s *UserService) ResolveRoleIDs(roleCodes []string) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	if len(roleCodes) == 0 {
		return nil, nil
	}
	type roleIDRow struct {
		ID   int64  `gorm:"column:Id"`
		Name string `gorm:"column:Name"`
	}
	var rows []roleIDRow
	err := s.db.Table(`"Identity_Roles"`).
		Select(`"Id", "Name"`).
		Where(`"Name" IN ?`, roleCodes).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("查询角色ID失败: %w", err)
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, fmt.Sprintf("%d", row.ID))
	}
	return result, nil
}

type UserCreateRequest struct {
	Username       string   `json:"username" binding:"required"`
	Password       string   `json:"password"`
	Email          string   `json:"email"`
	PhoneNumber    string   `json:"phoneNumber"`
	RealName       string   `json:"realName"`
	Gender         string   `json:"gender"`
	Type           string   `json:"type"`
	Sort           int      `json:"sort"`
	IdNumber       string   `json:"idNumber"`
	ICNumber       string   `json:"icNumber"`
	Avatar         string   `json:"avatar"`
	Birthdate      string   `json:"birthdate"`
	SignatureImage string   `json:"signatureImage"`
	Roles          []string `json:"roles"`
	CreatorID      int64    `json:"creatorId"`
}

type UserUpdateRequest struct {
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	PhoneNumber    string   `json:"phoneNumber"`
	RealName       string   `json:"realName"`
	Gender         string   `json:"gender"`
	Type           string   `json:"type"`
	Sort           *int     `json:"sort"`
	IdNumber       string   `json:"idNumber"`
	ICNumber       string   `json:"icNumber"`
	Avatar         string   `json:"avatar"`
	Birthdate      string   `json:"birthdate"`
	SignatureImage string   `json:"signatureImage"`
	Roles          []string `json:"roles"`
	CreatorID      int64    `json:"creatorId"`
}

type UserUpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
