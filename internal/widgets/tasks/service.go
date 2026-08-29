package tasks

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/dandydeveloper/dandy-dashboard/internal/store"
)

const bucket = "tasks"

type Service struct {
	store store.Store
}

func NewService(s store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) List() ([]Task, error) {
	keys, err := s.store.Keys(bucket)
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(keys))
	for _, key := range keys {
		data, err := s.store.Get(bucket, key)
		if err != nil || data == nil {
			continue
		}
		var t Task
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	// Active tasks first, then done; newest first within each group.
	sort.Slice(tasks, func(i, j int) bool {
		iDone := tasks[i].Status == StatusDone
		jDone := tasks[j].Status == StatusDone
		if iDone != jDone {
			return !iDone
		}
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (s *Service) Create(title, description string, priority Priority, dueDate string) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title required")
	}
	if priority == "" {
		priority = PriorityMedium
	}
	now := time.Now().UTC()
	t := Task{
		ID:          fmt.Sprintf("%d", now.UnixNano()),
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return &t, s.store.Set(bucket, t.ID, data)
}

func (s *Service) Update(id string, updates map[string]any) (*Task, error) {
	data, err := s.store.Get(bucket, id)
	if err != nil || data == nil {
		return nil, fmt.Errorf("task not found")
	}
	var t Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	if v, ok := updates["title"].(string); ok && v != "" {
		t.Title = v
	}
	if v, ok := updates["description"].(string); ok {
		t.Description = v
	}
	if v, ok := updates["status"].(string); ok {
		t.Status = Status(v)
	}
	if v, ok := updates["priority"].(string); ok {
		t.Priority = Priority(v)
	}
	if v, ok := updates["due_date"].(string); ok {
		t.DueDate = v
	}
	t.UpdatedAt = time.Now().UTC()
	out, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return &t, s.store.Set(bucket, t.ID, out)
}

func (s *Service) Delete(id string) error {
	return s.store.Delete(bucket, id)
}
