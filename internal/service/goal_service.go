package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/spendly/backend/internal/domain"
	"github.com/spendly/backend/internal/repository"
)

type GoalResponse struct {
	Goal     *domain.Goal `json:"goal"`
	ETA_Days int          `json:"eta_days"` // Days required to reach goal based on recent velocity
}

type GoalService interface {
	CreateGoal(ctx context.Context, title string, amount float64, targetDate time.Time) (*domain.Goal, error)
	GetAllGoals(ctx context.Context) ([]GoalResponse, error)
	ContributeToGoal(ctx context.Context, goalID string, amount float64, note string) (*domain.GoalContribution, error)
}

type goalService struct {
	goalRepo repository.GoalRepository
}

func NewGoalService(goalRepo repository.GoalRepository) GoalService {
	return &goalService{goalRepo: goalRepo}
}

func (s *goalService) CreateGoal(ctx context.Context, title string, amount float64, targetDate time.Time) (*domain.Goal, error) {
	if amount <= 0 {
		return nil, errors.New("target amount must be positive")
	}

	goal := &domain.Goal{
		Title:         title,
		TargetAmount:  amount,
		CurrentAmount: 0,
		TargetDate:    targetDate,
	}

	if err := s.goalRepo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *goalService) GetAllGoals(ctx context.Context) ([]GoalResponse, error) {
	goals, err := s.goalRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []GoalResponse
	for i := range goals {
		g := &goals[i]
		// Progressive Engine ETA Logic:
		// (Assume simple extrapolation: Current velocity since creation vs total amount)
		daysElapsed := time.Since(g.CreatedAt).Hours() / 24
		etaDays := 0
		if g.CurrentAmount > 0 && daysElapsed > 1 {
			avgPerDay := float64(g.CurrentAmount) / float64(daysElapsed)
			remaining := g.TargetAmount - g.CurrentAmount
			if remaining > 0 {
				etaDays = int(math.Ceil(remaining / avgPerDay))
			}
		} else if g.CurrentAmount >= g.TargetAmount {
			etaDays = 0 // Achieved
		} else {
			// Not enough data to predict, fallback to total difference vs target date
			etaDays = int(math.Ceil(time.Until(g.TargetDate).Hours() / 24))
		}

		if etaDays < 0 {
			etaDays = 0
		}

		responses = append(responses, GoalResponse{
			Goal:     g,
			ETA_Days: etaDays,
		})
	}
	return responses, nil
}

func (s *goalService) ContributeToGoal(ctx context.Context, goalID string, amount float64, note string) (*domain.GoalContribution, error) {
	if amount <= 0 {
		return nil, errors.New("contribution amount must be positive")
	}

	goal, err := s.goalRepo.FindByID(ctx, goalID)
	if err != nil {
		return nil, errors.New("goal not found")
	}

	// 1. Transaction to be created linking this contribution to spending/wealth logic (Goal = expense in tracking)
	tx := &domain.Transaction{
		Title:       "Contribution to: " + goal.Title,
		Amount:      amount,
		Type:        "goal", // Specialized transaction type
		Date:        time.Now(),
		Note:        note,
		IsRecurring: false,
	}

	// 2. GoalContribution Entry
	contribution := &domain.GoalContribution{
		GoalID: goalID,
		Amount: amount,
		Date:   time.Now(),
		Note:   note,
	}

	// Combine atomic insert logic in Repository layer
	if err := s.goalRepo.AddContribution(ctx, contribution, tx); err != nil {
		return nil, err
	}

	return contribution, nil
}
