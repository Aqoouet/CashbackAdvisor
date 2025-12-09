package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

// UserGroup хранит информацию о группе пользователя
type UserGroup struct {
	UserID    string `json:"user_id"`
	GroupName string `json:"group_name"`
}

// GroupsStore управляет группами пользователей
type GroupsStore struct {
	mu     sync.RWMutex
	users  map[string]string // userID -> groupName
	groups map[string]bool   // groupName -> exists
	file   string
}

// NewGroupsStore создает новое хранилище групп
func NewGroupsStore(filePath string) *GroupsStore {
	store := &GroupsStore{
		users:  make(map[string]string),
		groups: make(map[string]bool),
		file:   filePath,
	}
	store.load()
	return store
}

// load загружает данные из файла
func (s *GroupsStore) load() {
	data, err := ioutil.ReadFile(s.file)
	if err != nil {
		// Файл не существует - это нормально для первого запуска
		return
	}

	var users []UserGroup
	if err := json.Unmarshal(data, &users); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, ug := range users {
		s.users[ug.UserID] = ug.GroupName
		s.groups[ug.GroupName] = true
	}
}

// save сохраняет данные в файл
func (s *GroupsStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []UserGroup
	for userID, groupName := range s.users {
		users = append(users, UserGroup{
			UserID:    userID,
			GroupName: groupName,
		})
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.file, data, 0644)
}

// GetUserGroup возвращает группу пользователя
func (s *GroupsStore) GetUserGroup(userID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	group, exists := s.users[userID]
	return group, exists
}

// SetUserGroup устанавливает группу пользователя
func (s *GroupsStore) SetUserGroup(userID, groupName string) error {
	s.mu.Lock()
	s.users[userID] = groupName
	s.groups[groupName] = true
	s.mu.Unlock()
	return s.save()
}

// GroupExists проверяет существование группы
func (s *GroupsStore) GroupExists(groupName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.groups[groupName]
}

// CreateGroup создает новую группу
func (s *GroupsStore) CreateGroup(groupName string, creatorID string) error {
	s.mu.Lock()
	if s.groups[groupName] {
		s.mu.Unlock()
		return fmt.Errorf("группа '%s' уже существует", groupName)
	}
	s.groups[groupName] = true
	s.users[creatorID] = groupName
	s.mu.Unlock()
	return s.save()
}

// GetGroupMembers возвращает список пользователей в группе
func (s *GroupsStore) GetGroupMembers(groupName string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var members []string
	for userID, group := range s.users {
		if group == groupName {
			members = append(members, userID)
		}
	}
	return members
}

// GetAllGroups возвращает список всех групп
func (s *GroupsStore) GetAllGroups() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var groups []string
	for group := range s.groups {
		groups = append(groups, group)
	}
	return groups
}

