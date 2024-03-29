package servers

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"unsafe"

	"github.com/echo-marche/presence-api/models"
	pb "github.com/echo-marche/presence-api/proto/pb"
	"github.com/echo-marche/presence-api/services"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/gorp.v2"
)

type PresenceServer struct {
	DbMap *gorp.DbMap
}

func (server *PresenceServer) Healthz(ctx context.Context, _ *emptypb.Empty) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{ReturnMessage: "Presence Server healthz ok"}, nil
}

func (server *PresenceServer) GetUserList(ctx context.Context, req *pb.UserListRequest) (*pb.UserListResponse, error) {
	responseUsers := []*pb.UserResponse{}
	var users models.Users
	_, err := server.DbMap.Select(&users, "SELECT name, email FROM users")
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.DataLoss,
			err.Error())
	}
	for _, user := range users {
		responseUsers = append(responseUsers, &pb.UserResponse{Name: user.Name.String, Email: user.Email})
	}
	return &pb.UserListResponse{Users: responseUsers}, nil
}

func (server *PresenceServer) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	// systemCode := req.GetSystemCode
	return &pb.UserResponse{Name: "test1", Email: "test1@gmail.com"}, nil
}

func (server *PresenceServer) UserRegistration(ctx context.Context, req *pb.UserRegistrationRequest) (*pb.StatusResponse, error) {
	// TODO ユーザーチェック
	// create token
	tokenString := services.CreateTokenString(req)
	// password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.DataLoss,
			err.Error())
	}
	tempUser := models.User{
		Email:              req.Email,
		EncryptedPassword:  *(*string)(unsafe.Pointer(&passwordHash)),
		ConfirmationSentAt: sql.NullTime{Time: time.Now(), Valid: true},
		ConfirmationToken:  sql.NullString{String: tokenString, Valid: true},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	// DB登録
	err = server.DbMap.Insert(&tempUser)
	if err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.Unavailable,
			err.Error())
	}
	return &pb.StatusResponse{StatusCode: "ok presence!"}, nil
}
