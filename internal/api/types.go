package api

type Torrent struct {
	Hash             string  `json:"hash"`
	Name             string  `json:"name"`
	Size             int64   `json:"size"`
	Progress         float64 `json:"progress"`
	DlSpeed          int64   `json:"dlspeed"`
	UpSpeed          int64   `json:"upspeed"`
	Priority         int     `json:"priority"`
	NumSeeds         int     `json:"num_seeds"`
	NumLeeches       int     `json:"num_leechers"`
	NumComplete      int     `json:"num_complete"`
	NumIncomplete    int     `json:"num_incomplete"`
	Ratio            float64 `json:"ratio"`
	ETA              int64   `json:"eta"`
	State            string  `json:"state"`
	Category         string  `json:"category"`
	Tags             string  `json:"tags"`
	AddedOn          int64   `json:"added_on"`
	CompletedOn      int64   `json:"completion_on"`
	Tracker          string  `json:"tracker"`
	SavePath         string  `json:"save_path"`
	Downloaded       int64   `json:"downloaded"`
	Uploaded         int64   `json:"uploaded"`
	RemainingSize    int64   `json:"amount_left"`
	TimeActive       int64   `json:"time_active"`
	AutoTMM          bool    `json:"auto_tmm"`
	TotalSize        int64   `json:"total_size"`
	MaxRatio         float64 `json:"max_ratio"`
	MaxSeedingTime   int64   `json:"max_seeding_time"`
	SeedingTimeLimit int64   `json:"seeding_time_limit"`
}

type GlobalStats struct {
	DlInfoSpeed      int64  `json:"dl_info_speed"`
	UpInfoSpeed      int64  `json:"up_info_speed"`
	DlInfoData       int64  `json:"dl_info_data"`
	UpInfoData       int64  `json:"up_info_data"`
	ConnectionStatus string `json:"connection_status"`
	DHTNodes         int64  `json:"dht_nodes"`
	FreeSpaceOnDisk  int64  `json:"free_space_on_disk"` // From /api/v2/sync/maindata
}

// MainData represents the response from /api/v2/sync/maindata
type MainData struct {
	ServerState ServerState `json:"server_state"`
}

// ServerState contains server state information including free disk space
type ServerState struct {
	ConnectionStatus string `json:"connection_status"`
	DHTNodes         int64  `json:"dht_nodes"`
	DlInfoSpeed      int64  `json:"dl_info_speed"`
	UpInfoSpeed      int64  `json:"up_info_speed"`
	DlInfoData       int64  `json:"dl_info_data"`
	UpInfoData       int64  `json:"up_info_data"`
	FreeSpaceOnDisk  int64  `json:"free_space_on_disk"`
}

// SyncMainDataResponse represents the full response from /api/v2/sync/maindata
// This endpoint provides incremental updates to reduce bandwidth usage
type SyncMainDataResponse struct {
	RID               int                       `json:"rid"`                // Response ID for tracking incremental updates
	FullUpdate        bool                      `json:"full_update"`        // Whether this is a full update (true) or incremental (false)
	Torrents          map[string]PartialTorrent `json:"torrents"`           // Map of hash -> torrent data (only changed fields in incremental updates)
	TorrentsRemoved   []string                  `json:"torrents_removed"`   // List of torrent hashes removed since last request
	Categories        map[string]Category       `json:"categories"`         // Categories added/updated since last request
	CategoriesRemoved []string                  `json:"categories_removed"` // Categories removed since last request
	Tags              []string                  `json:"tags"`               // Tags added since last request
	TagsRemoved       []string                  `json:"tags_removed"`       // Tags removed since last request
	ServerState       ServerState               `json:"server_state"`       // Current server state (always included)
}

// PartialTorrent represents torrent data from sync/maindata incremental updates.
// Uses pointer types to distinguish between "field not present" (nil) and "field is zero value".
// This is essential because the sync API only sends changed fields in incremental updates.
type PartialTorrent struct {
	Hash             *string  `json:"hash"`
	Name             *string  `json:"name"`
	Size             *int64   `json:"size"`
	Progress         *float64 `json:"progress"`
	DlSpeed          *int64   `json:"dlspeed"`
	UpSpeed          *int64   `json:"upspeed"`
	Priority         *int     `json:"priority"`
	NumSeeds         *int     `json:"num_seeds"`
	NumLeeches       *int     `json:"num_leechers"`
	NumComplete      *int     `json:"num_complete"`
	NumIncomplete    *int     `json:"num_incomplete"`
	Ratio            *float64 `json:"ratio"`
	ETA              *int64   `json:"eta"`
	State            *string  `json:"state"`
	Category         *string  `json:"category"`
	Tags             *string  `json:"tags"`
	AddedOn          *int64   `json:"added_on"`
	CompletedOn      *int64   `json:"completion_on"`
	Tracker          *string  `json:"tracker"`
	SavePath         *string  `json:"save_path"`
	Downloaded       *int64   `json:"downloaded"`
	Uploaded         *int64   `json:"uploaded"`
	RemainingSize    *int64   `json:"amount_left"`
	TimeActive       *int64   `json:"time_active"`
	AutoTMM          *bool    `json:"auto_tmm"`
	TotalSize        *int64   `json:"total_size"`
	MaxRatio         *float64 `json:"max_ratio"`
	MaxSeedingTime   *int64   `json:"max_seeding_time"`
	SeedingTimeLimit *int64   `json:"seeding_time_limit"`
}

// ApplyTo merges the partial torrent data into an existing torrent.
// Only fields that are non-nil (i.e., were present in the JSON) are updated.
func (p *PartialTorrent) ApplyTo(t *Torrent) {
	if p.Hash != nil {
		t.Hash = *p.Hash
	}
	if p.Name != nil {
		t.Name = *p.Name
	}
	if p.Size != nil {
		t.Size = *p.Size
	}
	if p.Progress != nil {
		t.Progress = *p.Progress
	}
	if p.DlSpeed != nil {
		t.DlSpeed = *p.DlSpeed
	}
	if p.UpSpeed != nil {
		t.UpSpeed = *p.UpSpeed
	}
	if p.Priority != nil {
		t.Priority = *p.Priority
	}
	if p.NumSeeds != nil {
		t.NumSeeds = *p.NumSeeds
	}
	if p.NumLeeches != nil {
		t.NumLeeches = *p.NumLeeches
	}
	if p.NumComplete != nil {
		t.NumComplete = *p.NumComplete
	}
	if p.NumIncomplete != nil {
		t.NumIncomplete = *p.NumIncomplete
	}
	if p.Ratio != nil {
		t.Ratio = *p.Ratio
	}
	if p.ETA != nil {
		t.ETA = *p.ETA
	}
	if p.State != nil {
		t.State = *p.State
	}
	if p.Category != nil {
		t.Category = *p.Category
	}
	if p.Tags != nil {
		t.Tags = *p.Tags
	}
	if p.AddedOn != nil {
		t.AddedOn = *p.AddedOn
	}
	if p.CompletedOn != nil {
		t.CompletedOn = *p.CompletedOn
	}
	if p.Tracker != nil {
		t.Tracker = *p.Tracker
	}
	if p.SavePath != nil {
		t.SavePath = *p.SavePath
	}
	if p.Downloaded != nil {
		t.Downloaded = *p.Downloaded
	}
	if p.Uploaded != nil {
		t.Uploaded = *p.Uploaded
	}
	if p.RemainingSize != nil {
		t.RemainingSize = *p.RemainingSize
	}
	if p.TimeActive != nil {
		t.TimeActive = *p.TimeActive
	}
	if p.AutoTMM != nil {
		t.AutoTMM = *p.AutoTMM
	}
	if p.TotalSize != nil {
		t.TotalSize = *p.TotalSize
	}
	if p.MaxRatio != nil {
		t.MaxRatio = *p.MaxRatio
	}
	if p.MaxSeedingTime != nil {
		t.MaxSeedingTime = *p.MaxSeedingTime
	}
	if p.SeedingTimeLimit != nil {
		t.SeedingTimeLimit = *p.SeedingTimeLimit
	}
}

// ToTorrent converts a PartialTorrent to a full Torrent.
// Used for new torrents where all fields should be present.
func (p *PartialTorrent) ToTorrent() Torrent {
	t := Torrent{}
	p.ApplyTo(&t)
	return t
}

// Category represents a torrent category
type Category struct {
	Name         string `json:"name"`
	SavePath     string `json:"savePath"`
	DownloadPath string `json:"download_path"`
}

type TorrentProperties struct {
	SavePath               string  `json:"save_path"`
	CreationDate           int64   `json:"creation_date"`
	PieceSize              int64   `json:"piece_size"`
	Comment                string  `json:"comment"`
	TotalWasted            int64   `json:"total_wasted"`
	TotalUploaded          int64   `json:"total_uploaded"`
	TotalUploadedSession   int64   `json:"total_uploaded_session"`
	TotalDownloaded        int64   `json:"total_downloaded"`
	TotalDownloadedSession int64   `json:"total_downloaded_session"`
	UpLimit                int64   `json:"up_limit"`
	DlLimit                int64   `json:"dl_limit"`
	TimeElapsed            int64   `json:"time_elapsed"`
	SeedingTime            int64   `json:"seeding_time"`
	NbConnections          int     `json:"nb_connections"`
	NbConnectionsLimit     int     `json:"nb_connections_limit"`
	ShareRatio             float64 `json:"share_ratio"`
	AdditionDate           int64   `json:"addition_date"`
	CompletionDate         int64   `json:"completion_date"`
	CreatedBy              string  `json:"created_by"`
	DlSpeedAvg             int64   `json:"dl_speed_avg"`
	DlSpeed                int64   `json:"dl_speed"`
	Eta                    int64   `json:"eta"`
	LastSeen               int64   `json:"last_seen"`
	Peers                  int     `json:"peers"`
	PeersTotal             int     `json:"peers_total"`
	PiecesCompleted        int     `json:"pieces_completed"`
	PiecesNum              int     `json:"pieces_num"`
	Reannounce             int64   `json:"reannounce"`
	Seeds                  int     `json:"seeds"`
	SeedsTotal             int     `json:"seeds_total"`
	TotalSize              int64   `json:"total_size"`
	UpSpeedAvg             int64   `json:"up_speed_avg"`
	UpSpeed                int64   `json:"up_speed"`
}

// Tracker represents a torrent tracker
type Tracker struct {
	URL           string `json:"url"`
	Status        int    `json:"status"`
	Tier          int    `json:"tier"`
	NumPeers      int    `json:"num_peers"`
	NumSeeds      int    `json:"num_seeds"`
	NumLeeches    int    `json:"num_leeches"`
	NumDownloaded int    `json:"num_downloaded"`
	Msg           string `json:"msg"`
}

// Peer represents a torrent peer
type Peer struct {
	IP          string  `json:"ip"`
	Port        int     `json:"port"`
	Country     string  `json:"country"`
	Connection  string  `json:"connection"`
	Flags       string  `json:"flags"`
	Client      string  `json:"client"`
	Progress    float64 `json:"progress"`
	DlSpeed     int64   `json:"dl_speed"`
	UpSpeed     int64   `json:"up_speed"`
	Downloaded  int64   `json:"downloaded"`
	Uploaded    int64   `json:"uploaded"`
	Relevance   float64 `json:"relevance"`
	FilesString string  `json:"files"`
}

// TorrentFile represents a file within a torrent
type TorrentFile struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Progress     float64 `json:"progress"`
	Priority     int     `json:"priority"`
	IsSeed       bool    `json:"is_seed"`
	PieceRange   []int   `json:"piece_range"`
	Availability float64 `json:"availability"`
}

type TorrentState string

const (
	StateError              TorrentState = "error"
	StateMissingFiles       TorrentState = "missingFiles"
	StateUploading          TorrentState = "uploading"
	StatePausedUP           TorrentState = "pausedUP"
	StateStoppedUP          TorrentState = "stoppedUP" // Undocumented but real state
	StateQueuedUP           TorrentState = "queuedUP"
	StateStalledUP          TorrentState = "stalledUP"
	StateForcedUP           TorrentState = "forcedUP"
	StateAllocating         TorrentState = "allocating"
	StateDownloading        TorrentState = "downloading"
	StateMetaDL             TorrentState = "metaDL"
	StatePausedDL           TorrentState = "pausedDL"
	StateQueuedDL           TorrentState = "queuedDL"
	StateStalledDL          TorrentState = "stalledDL"
	StateForcedDL           TorrentState = "forcedDL"
	StateCheckingDL         TorrentState = "checkingDL"
	StateCheckingUP         TorrentState = "checkingUP"
	StateQueuedForChecking  TorrentState = "queuedForChecking"
	StateCheckingResumeData TorrentState = "checkingResumeData"
	StateMoving             TorrentState = "moving"
	StateUnknown            TorrentState = "unknown"
)

func (s TorrentState) String() string {
	return string(s)
}

func (s TorrentState) IsDownloading() bool {
	switch s {
	case StateDownloading, StateMetaDL, StateForcedDL, StateAllocating:
		return true
	default:
		return false
	}
}

func (s TorrentState) IsUploading() bool {
	switch s {
	case StateUploading, StateForcedUP, StateStalledUP:
		return true
	default:
		return false
	}
}

func (s TorrentState) IsPaused() bool {
	return s == StatePausedDL || s == StatePausedUP || s == StateStoppedUP
}

func (s TorrentState) IsActive() bool {
	return s.IsDownloading() || s.IsUploading()
}
