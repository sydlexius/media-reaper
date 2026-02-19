package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/sydlexius/media-reaper/internal/config"
	"github.com/sydlexius/media-reaper/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionName    = "media-reaper-session"
	sessionUserKey = "user_id"
	bcryptCost     = 12
)

type Service struct {
	users repository.UserRepository
	store sessions.Store
	cfg   *config.Config
}

func NewService(users repository.UserRepository, cfg *config.Config) *Service {
	secret := cfg.SessionSecret
	if secret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("Failed to generate session secret: %v", err)
		}
		secret = hex.EncodeToString(b)
		log.Println("WARNING: No session secret configured. Generated a random one. Sessions will not persist across restarts.")
	}

	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.SecureCookies,
	}

	return &Service{
		users: users,
		store: store,
		cfg:   cfg,
	}
}

func (s *Service) Bootstrap(ctx context.Context) error {
	if s.cfg.AdminUser == "" || s.cfg.AdminPass == "" {
		return nil
	}

	count, err := s.users.Count(ctx)
	if err != nil {
		return fmt.Errorf("checking user count: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPass), bcryptCost)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}

	user := &repository.User{
		ID:           uuid.New().String(),
		Username:     s.cfg.AdminUser,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	if err := s.users.Create(ctx, user); err != nil {
		return fmt.Errorf("creating admin user: %w", err)
	}

	log.Printf("Admin user '%s' created during bootstrap", s.cfg.AdminUser)
	return nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*repository.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return user, nil
}

func (s *Service) GetSession(r *http.Request) (*sessions.Session, error) {
	return s.store.Get(r, sessionName)
}

func (s *Service) GetUserFromSession(ctx context.Context, r *http.Request) (*repository.User, error) {
	sess, err := s.store.Get(r, sessionName)
	if err != nil {
		return nil, err
	}

	userID, ok := sess.Values[sessionUserKey].(string)
	if !ok || userID == "" {
		return nil, nil
	}

	return s.users.GetByID(ctx, userID)
}

func (s *Service) CreateSession(w http.ResponseWriter, r *http.Request, userID string) error {
	sess, err := s.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	sess.Values[sessionUserKey] = userID
	return sess.Save(r, w)
}

func (s *Service) DestroySession(w http.ResponseWriter, r *http.Request) error {
	sess, err := s.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}
