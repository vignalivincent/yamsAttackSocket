package game

import (
	"time"
)

func (s *ServerStats) GetUptime() time.Duration {
    s.Mutex.RLock()
    defer s.Mutex.RUnlock()
    return time.Since(s.StartTime)
}

func (s *ServerStats) FormatResponse() ServerStatsResponse {
    uptime := s.GetUptime()
    
    s.Mutex.RLock()
    defer s.Mutex.RUnlock()
    
    return ServerStatsResponse{
        TotalGamesCreated:    s.TotalGamesCreated,
        ActiveGames:          s.ActiveGames,
        TotalViewers:         s.TotalViewers,
        TotalHostConnections: s.TotalHostConnections,
        Uptime:               uptime.String(),
        StartTime:            s.StartTime.Format(time.RFC3339),
    }
}