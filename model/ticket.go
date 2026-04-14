package model

import (
	"errors"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	TicketStatusOpen       = 1 // 待处理
	TicketStatusProcessing = 2 // 处理中
	TicketStatusResolved   = 3 // 已解决
	TicketStatusClosed     = 4 // 已关闭
)

const (
	TicketTypeGeneral = "general"
	TicketTypeRefund  = "refund"
	TicketTypeInvoice = "invoice"
)

var (
	ErrTicketNotFound              = errors.New("ticket not found")
	ErrTicketForbidden             = errors.New("ticket forbidden")
	ErrTicketClosed                = errors.New("ticket closed")
	ErrTicketSubjectEmpty          = errors.New("ticket subject empty")
	ErrTicketContentEmpty          = errors.New("ticket content empty")
	ErrTicketInvalidStatus         = errors.New("ticket invalid status")
	ErrTicketInvalidType           = errors.New("ticket invalid type")
	ErrTicketInvoiceNotFound       = errors.New("ticket invoice not found")
	ErrTicketInvoiceStatusInvalid  = errors.New("ticket invoice status invalid")
	ErrTicketInvoiceOrderEmpty     = errors.New("ticket invoice order empty")
	ErrTicketInvoiceOrderInvalid   = errors.New("ticket invoice order invalid")
	ErrTicketInvoiceOrderDuplicate = errors.New("ticket invoice order duplicate")
	ErrTicketInvoiceCompanyEmpty   = errors.New("ticket invoice company empty")
	ErrTicketInvoiceTaxNumberEmpty = errors.New("ticket invoice tax number empty")
	ErrTicketInvoiceEmailEmpty     = errors.New("ticket invoice email empty")
)

type Ticket struct {
	Id          int            `json:"id"`
	UserId      int            `json:"user_id" gorm:"index;not null"`
	Username    string         `json:"username" gorm:"type:varchar(64)"`
	Subject     string         `json:"subject" gorm:"type:varchar(255);not null"`
	Type        string         `json:"type" gorm:"type:varchar(32);index;default:'general'"`
	Status      int            `json:"status" gorm:"type:int;index;default:1"`
	Priority    int            `json:"priority" gorm:"type:int;default:2"`
	AdminId     int            `json:"admin_id" gorm:"type:int;default:0"`
	CreatedTime int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime int64          `json:"updated_time" gorm:"bigint;index"`
	ClosedTime  int64          `json:"closed_time" gorm:"bigint;default:0"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type TicketQueryOptions struct {
	UserId  int
	Status  int
	Type    string
	Keyword string
}

type CreateTicketParams struct {
	UserId   int
	Username string
	Subject  string
	Type     string
	Priority int
	Content  string
	Role     int
}

func (ticket *Ticket) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if ticket.CreatedTime == 0 {
		ticket.CreatedTime = now
	}
	if ticket.UpdatedTime == 0 {
		ticket.UpdatedTime = now
	}
	return nil
}

func (ticket *Ticket) BeforeUpdate(tx *gorm.DB) error {
	ticket.UpdatedTime = common.GetTimestamp()
	return nil
}

func NormalizeTicketType(ticketType string) string {
	return strings.ToLower(strings.TrimSpace(ticketType))
}

func NormalizeTicketPriority(priority int) int {
	switch priority {
	case 1, 2, 3:
		return priority
	default:
		return 2
	}
}

func IsValidTicketType(ticketType string) bool {
	switch NormalizeTicketType(ticketType) {
	case TicketTypeGeneral, TicketTypeRefund, TicketTypeInvoice:
		return true
	default:
		return false
	}
}

func IsValidTicketStatus(status int) bool {
	switch status {
	case TicketStatusOpen, TicketStatusProcessing, TicketStatusResolved, TicketStatusClosed:
		return true
	default:
		return false
	}
}

func CreateTicketWithMessage(params CreateTicketParams) (*Ticket, *TicketMessage, error) {
	subject := strings.TrimSpace(params.Subject)
	if subject == "" {
		return nil, nil, ErrTicketSubjectEmpty
	}
	content := strings.TrimSpace(params.Content)
	if content == "" {
		return nil, nil, ErrTicketContentEmpty
	}

	ticketType := NormalizeTicketType(params.Type)
	if ticketType == "" {
		ticketType = TicketTypeGeneral
	}
	if !IsValidTicketType(ticketType) {
		return nil, nil, ErrTicketInvalidType
	}

	now := common.GetTimestamp()
	ticket := &Ticket{
		UserId:      params.UserId,
		Username:    strings.TrimSpace(params.Username),
		Subject:     subject,
		Type:        ticketType,
		Status:      TicketStatusOpen,
		Priority:    NormalizeTicketPriority(params.Priority),
		CreatedTime: now,
		UpdatedTime: now,
	}
	message := &TicketMessage{
		UserId:      params.UserId,
		Username:    strings.TrimSpace(params.Username),
		Role:        params.Role,
		Content:     content,
		CreatedTime: now,
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		message.TicketId = ticket.Id
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return ticket, message, nil
}

func applyTicketFilters(query *gorm.DB, options TicketQueryOptions) *gorm.DB {
	if options.UserId > 0 {
		query = query.Where("user_id = ?", options.UserId)
	}
	if options.Status > 0 {
		query = query.Where("status = ?", options.Status)
	}
	ticketType := NormalizeTicketType(options.Type)
	if ticketType != "" {
		query = query.Where("type = ?", ticketType)
	}
	keyword := strings.TrimSpace(options.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		if ticketId, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("(id = ? OR subject LIKE ? OR username LIKE ?)", ticketId, like, like)
		} else {
			query = query.Where("(subject LIKE ? OR username LIKE ?)", like, like)
		}
	}
	return query
}

func ListTickets(options TicketQueryOptions, pageInfo *common.PageInfo) (tickets []*Ticket, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := applyTicketFilters(tx.Model(&Ticket{}), options)
	if err = query.Count(&total).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	if err = query.Order("updated_time desc, id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&tickets).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
}

func GetTicketById(ticketId int) (*Ticket, error) {
	var ticket Ticket
	if err := DB.First(&ticket, "id = ?", ticketId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

func GetUserTicketById(ticketId int, userId int) (*Ticket, error) {
	ticket, err := GetTicketById(ticketId)
	if err != nil {
		return nil, err
	}
	if ticket.UserId != userId {
		return nil, ErrTicketForbidden
	}
	return ticket, nil
}

func GetTicketMessages(ticketId int) (messages []*TicketMessage, err error) {
	err = DB.Where("ticket_id = ?", ticketId).Order("id asc").Find(&messages).Error
	return messages, err
}

func AddTicketMessage(ticketId int, userId int, username string, role int, content string) (*TicketMessage, *Ticket, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil, ErrTicketContentEmpty
	}

	var ticket Ticket
	now := common.GetTimestamp()
	message := &TicketMessage{
		TicketId:    ticketId,
		UserId:      userId,
		Username:    strings.TrimSpace(username),
		Role:        role,
		Content:     content,
		CreatedTime: now,
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ticket, "id = ?", ticketId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		if ticket.Status == TicketStatusClosed {
			return ErrTicketClosed
		}
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// 管理员首次接手后自动进入处理中，用户在已解决状态追问时也会回到处理中。
		updates := map[string]interface{}{
			"updated_time": now,
		}
		if role >= common.RoleAdminUser {
			updates["admin_id"] = userId
			ticket.AdminId = userId
			if ticket.Status == TicketStatusOpen {
				updates["status"] = TicketStatusProcessing
				ticket.Status = TicketStatusProcessing
			}
		} else if ticket.Status == TicketStatusResolved {
			updates["status"] = TicketStatusProcessing
			ticket.Status = TicketStatusProcessing
		}
		ticket.UpdatedTime = now

		if err := tx.Model(&Ticket{}).Where("id = ?", ticket.Id).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return message, &ticket, nil
}

func CloseUserTicket(ticketId int, userId int) (*Ticket, error) {
	var ticket Ticket
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ticket, "id = ?", ticketId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		if ticket.UserId != userId {
			return ErrTicketForbidden
		}
		if ticket.Status == TicketStatusClosed {
			return nil
		}

		now := common.GetTimestamp()
		if err := tx.Model(&Ticket{}).Where("id = ?", ticket.Id).Updates(map[string]interface{}{
			"status":       TicketStatusClosed,
			"closed_time":  now,
			"updated_time": now,
		}).Error; err != nil {
			return err
		}
		ticket.Status = TicketStatusClosed
		ticket.ClosedTime = now
		ticket.UpdatedTime = now
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func UpdateTicketStatus(ticketId int, adminId int, status *int, priority *int) (*Ticket, error) {
	var ticket Ticket
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ticket, "id = ?", ticketId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		now := common.GetTimestamp()
		updates := map[string]interface{}{
			"updated_time": now,
			"admin_id":     adminId,
		}
		ticket.AdminId = adminId
		ticket.UpdatedTime = now

		if status != nil {
			if !IsValidTicketStatus(*status) {
				return ErrTicketInvalidStatus
			}
			updates["status"] = *status
			ticket.Status = *status
			if *status == TicketStatusClosed {
				updates["closed_time"] = now
				ticket.ClosedTime = now
			} else {
				updates["closed_time"] = 0
				ticket.ClosedTime = 0
			}
		}
		if priority != nil {
			nextPriority := NormalizeTicketPriority(*priority)
			updates["priority"] = nextPriority
			ticket.Priority = nextPriority
		}

		if err := tx.Model(&Ticket{}).Where("id = ?", ticket.Id).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}
