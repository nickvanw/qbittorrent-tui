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
	NumTorrents      int    `json:"num_torrents"`
	NumActiveItems   int    `json:"num_active_torrents"`
	ConnectionStatus string `json:"connection_status"`
	DHT              bool   `json:"dht"`
	PeerExchange     bool   `json:"peer_exchange"`
	DHTNodes         int64  `json:"dht_nodes"`
	TorrentsCount    int    `json:"torrents_count"`
	FreeSpaceOnDisk  int64  `json:"free_space_on_disk"`
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
