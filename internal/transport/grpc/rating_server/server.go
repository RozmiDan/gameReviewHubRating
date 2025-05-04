package rating_server

import (
	"context"
	"errors"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	ratingv1 "github.com/RozmiDan/gamehub-protos/gen/go/gamehub"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RatingUseCase interface {
	SubmitRating(ctx context.Context, userID string, gameID string, rating int32) error
	GetGameRating(ctx context.Context, gameID string) (entity.GameRating, error)
	GetTopGames(ctx context.Context, limit, offset int32) ([]entity.GameRating, error)
}

type serverAPI struct {
	ratingv1.UnimplementedRatingServiceServer
	usecase RatingUseCase
}

func Register(grpcServer *grpc.Server, uc RatingUseCase) {
	ratingv1.RegisterRatingServiceServer(grpcServer, &serverAPI{usecase: uc})
}

func (s *serverAPI) SubmitRating(ctx context.Context,
	req *ratingv1.SubmitRatingRequest) (*ratingv1.SubmitRatingResponse, error) {

	if req.GetGameId() == "" || req.GetUserId() == "" || req.GetRating() == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid entered data")
	}

	if !validateUUID(req.GetUserId(), req.GetGameId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid entered uuid")
	}

	if err := s.usecase.SubmitRating(ctx, req.UserId, req.GameId, req.Rating); err != nil {
		return &ratingv1.SubmitRatingResponse{Success: false}, status.Error(codes.Internal, "could not submit rating")
	}

	return &ratingv1.SubmitRatingResponse{Success: true}, nil

}

func (s *serverAPI) GetGameRating(ctx context.Context,
	req *ratingv1.GetGameRatingRequest) (*ratingv1.GetGameRatingResponse, error) {

	if req.GetGameId() == "" {
		return &ratingv1.GetGameRatingResponse{}, status.Error(codes.InvalidArgument, "invalid gameID")
	}

	if !validateUUID(req.GetGameId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid entered uuid")
	}

	resEnt, err := s.usecase.GetGameRating(ctx, req.GameId)

	if err != nil {
		if errors.Is(err, entity.ErrGameNotFound) {
			return &ratingv1.GetGameRatingResponse{}, status.Error(codes.NotFound, "gameID not found")
		}
		return &ratingv1.GetGameRatingResponse{}, status.Error(codes.Internal, "internal error")
	}

	return &ratingv1.GetGameRatingResponse{
		GameId:        resEnt.GameId,
		AverageRating: float64(resEnt.AverageRating),
		RatingsCount:  resEnt.RatingsCount,
	}, nil
}

func (s *serverAPI) GetTopGames(ctx context.Context,
	req *ratingv1.GetTopGamesRequest) (*ratingv1.GetTopGamesResponse, error) {

	if req.GetLimit() != 10 {
		return &ratingv1.GetTopGamesResponse{}, status.Error(codes.InvalidArgument, "invalid limit")
	}

	list, err := s.usecase.GetTopGames(ctx, req.Limit, req.Offset)

	if err != nil {
		return nil, status.Error(codes.Internal, "could not get top games")
	}

	resp := &ratingv1.GetTopGamesResponse{}
	for _, e := range list {
		resp.Games = append(resp.Games, &ratingv1.GameRating{
			GameId:        e.GameId,
			AverageRating: float64(e.AverageRating),
			RatingsCount:  e.RatingsCount,
		})
	}
	return resp, nil
}

func validateUUID(uuids ...string) bool {
	for _, it := range uuids {
		if err := uuid.Validate(it); err != nil {
			return false
		}
	}
	return true
}
