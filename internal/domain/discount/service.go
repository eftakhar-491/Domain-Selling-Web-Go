package discount

import (
	"errors"
	"math"
	"strings"
	"time"

	"project-setup/internal/models"
)

type DiscountService struct {
	repo *DiscountRepository
}

func NewDiscountService(repo *DiscountRepository) *DiscountService {
	return &DiscountService{repo: repo}
}

func (s *DiscountService) Create(req CreateDiscountRequest) (*models.Discount, error) {
	if req.Scope == "TLD" && (req.TargetTLD == nil || *req.TargetTLD == "") {
		return nil, errors.New("target_tld is required for TLD discounts")
	}
	if req.Scope == "COUPON" && (req.Code == nil || *req.Code == "") {
		return nil, errors.New("code is required for coupon discounts")
	}

	// Normalize TLD (remove leading dot)
	if req.TargetTLD != nil {
		tld := strings.TrimPrefix(strings.ToLower(*req.TargetTLD), ".")
		req.TargetTLD = &tld
	}

	// Uppercase coupon code
	if req.Code != nil {
		code := strings.ToUpper(strings.TrimSpace(*req.Code))
		req.Code = &code
	}

	discount := &models.Discount{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		Type:              models.DiscountType(req.Type),
		Scope:             models.DiscountScope(req.Scope),
		TargetTLD:         req.TargetTLD,
		Value:             req.Value,
		MinSpend:          req.MinSpend,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		IsActive:          true,
	}

	if err := s.repo.Create(discount); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("discount code already exists")
		}
		return nil, errors.New("failed to create discount")
	}

	return discount, nil
}

func (s *DiscountService) GetByID(id uint) (*models.Discount, error) {
	discount, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("discount not found")
	}
	return discount, nil
}

func (s *DiscountService) GetAll(page, limit int) ([]models.Discount, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.FindAll(page, limit)
}

func (s *DiscountService) GetActiveTLDDiscounts() ([]models.Discount, error) {
	return s.repo.FindAllActiveTLD()
}

func (s *DiscountService) Update(id uint, req UpdateDiscountRequest) (*models.Discount, error) {
	discount, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("discount not found")
	}

	if req.Name != nil {
		discount.Name = *req.Name
	}
	if req.Description != nil {
		discount.Description = *req.Description
	}
	if req.Value != nil {
		discount.Value = *req.Value
	}
	if req.MinSpend != nil {
		discount.MinSpend = *req.MinSpend
	}
	if req.MaxDiscountAmount != nil {
		discount.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.UsageLimit != nil {
		discount.UsageLimit = *req.UsageLimit
	}
	if req.StartDate != nil {
		discount.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		discount.EndDate = req.EndDate
	}
	if req.IsActive != nil {
		discount.IsActive = *req.IsActive
	}

	if err := s.repo.Update(discount); err != nil {
		return nil, errors.New("failed to update discount")
	}

	return discount, nil
}

func (s *DiscountService) Delete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("discount not found")
	}
	return s.repo.Delete(id)
}

// ValidateCoupon checks if a coupon code is valid and returns the discount
func (s *DiscountService) ValidateCoupon(code string, cartTotal float64) (*models.Discount, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	discount, err := s.repo.FindByCode(code)
	if err != nil {
		return nil, errors.New("invalid coupon code")
	}

	if !discount.IsActive {
		return nil, errors.New("this coupon is no longer active")
	}

	if discount.Scope != models.DiscountScopeCoupon {
		return nil, errors.New("invalid coupon code")
	}

	now := time.Now()
	if discount.StartDate != nil && now.Before(*discount.StartDate) {
		return nil, errors.New("this coupon is not yet active")
	}
	if discount.EndDate != nil && now.After(*discount.EndDate) {
		return nil, errors.New("this coupon has expired")
	}

	if discount.UsageLimit > 0 && discount.UsedCount >= discount.UsageLimit {
		return nil, errors.New("this coupon has reached its usage limit")
	}

	if cartTotal < discount.MinSpend {
		return nil, errors.New("minimum spend not met for this coupon")
	}

	return discount, nil
}

// CalculateItemDiscount calculates the discount amount for a single cart item based on its TLD
func (s *DiscountService) CalculateItemDiscount(tld string, originalPrice float64) (float64, string) {
	tld = strings.TrimPrefix(strings.ToLower(tld), ".")

	discount, err := s.repo.FindActiveByTLD(tld)
	if err != nil {
		return 0, ""
	}

	amount := calculateDiscountAmount(discount, originalPrice)
	return amount, discount.Name
}

// CalculateCouponDiscount calculates the discount amount for a coupon on the cart total
func (s *DiscountService) CalculateCouponDiscount(discount *models.Discount, cartTotal float64) float64 {
	return calculateDiscountAmount(discount, cartTotal)
}

func calculateDiscountAmount(discount *models.Discount, base float64) float64 {
	var amount float64

	if discount.Type == models.DiscountTypePercentage {
		amount = base * (discount.Value / 100)
	} else {
		amount = discount.Value
	}

	// Cap at max discount
	if discount.MaxDiscountAmount != nil && amount > *discount.MaxDiscountAmount {
		amount = *discount.MaxDiscountAmount
	}

	// Don't exceed the base price
	amount = math.Min(amount, base)
	amount = math.Round(amount*100) / 100

	return amount
}
