package tournament

import (
	"errors"
	"testing"

	"github.com/OndasAlikhan/tourik/internal/domain"
)

func participants(n int) []domain.Participant {
	pts := make([]domain.Participant, n)
	for i := range pts {
		pts[i] = domain.Participant{ID: i + 1}
	}
	return pts
}

func TestTournament_CheckConditions(t *testing.T) {
	tests := []struct {
		name    string
		tour    Tournament
		pts     []domain.Participant
		wantErr error
	}{
		{
			name:    "fewer than two participants",
			tour:    Tournament{BracketFormat: SingleElimination, MaxParticipants: 4},
			pts:     participants(1),
			wantErr: ErrMaxParticipants,
		},
		{
			name:    "no participants",
			tour:    Tournament{BracketFormat: SingleElimination, MaxParticipants: 4},
			pts:     nil,
			wantErr: ErrMaxParticipants,
		},
		{
			name:    "single elimination with non power of two max participants",
			tour:    Tournament{BracketFormat: SingleElimination, MaxParticipants: 3},
			pts:     participants(3),
			wantErr: ErrMaxParticipants,
		},
		{
			name:    "double elimination with non power of two max participants",
			tour:    Tournament{BracketFormat: DoubleElimination, MaxParticipants: 6},
			pts:     participants(2),
			wantErr: ErrMaxParticipants,
		},
		{
			name:    "single elimination with power of two max participants",
			tour:    Tournament{BracketFormat: SingleElimination, MaxParticipants: 4},
			pts:     participants(2),
			wantErr: nil,
		},
		{
			name:    "round robin ignores power of two requirement",
			tour:    Tournament{BracketFormat: RoundRobin, MaxParticipants: 3},
			pts:     participants(3),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tour.CheckConditions(tt.pts)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error to wrap %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTournament_BeginRegistration(t *testing.T) {
	tests := []struct {
		name    string
		status  TournamentStatus
		wantErr bool
	}{
		{name: "from draft succeeds", status: Draft, wantErr: false},
		{name: "from registration fails", status: Registration, wantErr: true},
		{name: "from in progress fails", status: InProgress, wantErr: true},
		{name: "from finished fails", status: Finished, wantErr: true},
		{name: "from cancelled fails", status: Cancelled, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tour := Tournament{Status: tt.status}
			err := tour.BeginRegistration()

			if tt.wantErr {
				if !errors.Is(err, ErrWrongStatus) {
					t.Fatalf("expected ErrWrongStatus, got %v", err)
				}
				if tour.Status != tt.status {
					t.Fatalf("status should be unchanged on error, got %s", tour.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tour.Status != Registration {
				t.Fatalf("expected status %s, got %s", Registration, tour.Status)
			}
		})
	}
}

func TestTournament_Start(t *testing.T) {
	tests := []struct {
		name    string
		status  TournamentStatus
		wantErr bool
	}{
		{name: "from registration succeeds", status: Registration, wantErr: false},
		{name: "from draft fails", status: Draft, wantErr: true},
		{name: "from in progress fails", status: InProgress, wantErr: true},
		{name: "from finished fails", status: Finished, wantErr: true},
		{name: "from cancelled fails", status: Cancelled, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tour := Tournament{Status: tt.status}
			err := tour.Start()

			if tt.wantErr {
				if !errors.Is(err, ErrWrongStatus) {
					t.Fatalf("expected ErrWrongStatus, got %v", err)
				}
				if tour.Status != tt.status {
					t.Fatalf("status should be unchanged on error, got %s", tour.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tour.Status != InProgress {
				t.Fatalf("expected status %s, got %s", InProgress, tour.Status)
			}
		})
	}
}

func TestTournament_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		status  TournamentStatus
		wantErr bool
	}{
		{name: "from draft succeeds", status: Draft, wantErr: false},
		{name: "from registration succeeds", status: Registration, wantErr: false},
		{name: "from in progress succeeds", status: InProgress, wantErr: false},
		{name: "from finished fails", status: Finished, wantErr: true},
		{name: "from cancelled fails", status: Cancelled, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tour := Tournament{Status: tt.status}
			err := tour.Cancel()

			if tt.wantErr {
				if !errors.Is(err, ErrWrongStatus) {
					t.Fatalf("expected ErrWrongStatus, got %v", err)
				}
				if tour.Status != tt.status {
					t.Fatalf("status should be unchanged on error, got %s", tour.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tour.Status != Cancelled {
				t.Fatalf("expected status %s, got %s", Cancelled, tour.Status)
			}
		})
	}
}

func TestIsPowerOfTwo(t *testing.T) {
	tests := []struct {
		n    int
		want bool
	}{
		{n: -4, want: false},
		{n: -1, want: false},
		{n: 0, want: false},
		{n: 1, want: true},
		{n: 2, want: true},
		{n: 3, want: false},
		{n: 4, want: true},
		{n: 5, want: false},
		{n: 1024, want: true},
		{n: 1023, want: false},
	}

	for _, tt := range tests {
		if got := isPowerOfTwo(tt.n); got != tt.want {
			t.Errorf("isPowerOfTwo(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}
