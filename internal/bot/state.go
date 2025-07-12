package bot

import (
	"sync"

	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/openrouter"
)

// UserContext хранит контекст диалога для конкретного пользователя
type UserContext struct {
	Messages []openrouter.Message
}

// State управляет состоянием бота для всех пользователей
type State struct {
	userContexts map[int64]*UserContext
	maxHistory   int
	mu           sync.Mutex
}

func NewState(maxHistory int) *State {
	return &State{
		userContexts: make(map[int64]*UserContext),
		maxHistory:   maxHistory,
	}
}

func (s *State) AddMessage(userID int64, role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.userContexts[userID]; !ok {
		s.userContexts[userID] = &UserContext{
			Messages: make([]openrouter.Message, 0),
		}
	}

	s.userContexts[userID].Messages = append(s.userContexts[userID].Messages, openrouter.Message{
		Role:    role,
		Content: content,
	})

	if len(s.userContexts[userID].Messages) > s.maxHistory {
		s.userContexts[userID].Messages = s.userContexts[userID].Messages[len(s.userContexts[userID].Messages)-s.maxHistory:]
	}
}

func (s *State) GetMessages(userID int64) []openrouter.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ctx, ok := s.userContexts[userID]; ok {
		return ctx.Messages
	}
	return []openrouter.Message{}
}

func (s *State) ResetUser(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.userContexts, userID)
}
