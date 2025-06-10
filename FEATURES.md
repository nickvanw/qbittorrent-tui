# Feature Roadmap

This document outlines planned enhancements to qBittorrent TUI based on user feedback and UX improvements.

## ✅ All Features Completed!

All three requested features have been successfully implemented with excellent integration:

### Dynamic Sorting with Configurable Columns
The sorting system now intelligently adapts to user-configured columns:
- **Dynamic key mapping**: Number keys (1-9) map to the first 9 visible columns
- **Adaptive shortcuts**: Hide a column, and its sort shortcut transfers to the next visible column
- **Universal sorting**: All columns (including new ones like ETA, Category, Tags) are sortable
- **Visual feedback**: Sort indicators (↑/↓) appear in headers for the active sort column
- **Consistent behavior**: Sort direction toggle and secondary sorting work across all column configurations

## 🎯 Proposed Features

### 1. Responsive Layout - Full Terminal Width Usage
**Priority: High** | **Effort: Medium** | **Impact: High**

#### Current Issue
The UI uses fixed column widths and may not efficiently utilize available terminal width, especially on wide monitors.

#### Proposed Solution
- **Dynamic column sizing** - Columns expand/contract based on terminal width
- **Minimum/maximum width constraints** - Ensure readability on all screen sizes
- **Smart text truncation** - Intelligent ellipsis placement for long torrent names
- **Responsive breakpoints** - Hide less important columns on narrow terminals

#### Current Implementation Analysis
```go
// internal/ui/components/torrent_list.go (current)
var columns = []Column{
    {Title: "Name", Width: 40},      // Fixed 40 chars - TOO RIGID
    {Title: "Size", Width: 10},      // Fixed 10 chars - OK
    {Title: "Progress", Width: 10},  // Fixed 10 chars - OK
    // ... total: ~114 chars fixed width
}
```

**Problem**: Uses only 114 characters even on 200+ character terminals.

#### Implementation Details
```go
type ColumnConfig struct {
    Key         string  // "name", "size", "progress", etc.
    Title       string  // Display title
    MinWidth    int     // Minimum column width
    MaxWidth    int     // Maximum column width (0 = unlimited)
    FlexGrow    float64 // How much this column should grow (0-1)
    Priority    int     // Hide order on narrow screens (1=highest priority)
    Formatter   func(api.Torrent) string // Extract/format data
}

// Replace global columns variable with:
func (t *TorrentList) calculateColumnWidths(availableWidth int) []ColumnConfig {
    // Dynamic width calculation based on terminal size
}
```

#### UX Considerations
- Maintain readability on 80-column terminals
- Prioritize most important columns (name, progress, state)
- Smooth transitions when resizing terminal

---

### 2. Torrent Sorting
**Priority: High** | **Effort: Medium** | **Impact: High**

#### Current Issue
Torrents are displayed in API return order with no user control over sorting.

#### Proposed Solution
- **Multi-column sorting** - Sort by any visible column
- **Sort direction toggle** - Ascending/descending with visual indicators
- **Keyboard shortcuts** - Quick sort without leaving torrent list
- **Persistent preferences** - Remember user's preferred sort order

#### Implementation Details
```go
type SortConfig struct {
    Column    string    // "name", "size", "progress", "added_on", etc.
    Direction SortDir   // Ascending, Descending
    Secondary string    // Secondary sort column (for ties)
}

type SortDir int
const (
    SortAsc SortDir = iota
    SortDesc
)
```

#### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `1` | Sort by name |
| `2` | Sort by size |
| `3` | Sort by progress |
| `4` | Sort by state |
| `5` | Sort by speed |
| `6` | Sort by added date |
| `Shift+[1-6]` | Reverse sort direction |

#### Visual Indicators
```
Name ↑   Size     Progress  State     DL Speed  Added On
```

---

### 3. Configurable Columns
**Priority: Medium** | **Effort: High** | **Impact: High**

#### Current Issue
Only shows basic columns (Name, Size, Progress, State, Seeds, Peers, Speeds, Ratio). qBittorrent API provides much more data.

#### Available Data from API
From `internal/api/types.go`, we can add:
- **Dates**: Added On, Completion Date, Last Activity
- **Advanced Stats**: ETA, Time Active, Seeding Time
- **Metadata**: Category, Tags, Tracker, Save Path
- **Ratios**: Share Ratio, Upload Ratio
- **Technical**: Priority, Sequential Download, Piece Count

#### Proposed Solution
- **Column configuration UI** - In-app column selection
- **Configuration persistence** - Save to config file
- **Preset configurations** - "Basic", "Advanced", "Developer" presets
- **Runtime column toggle** - Show/hide columns without restart

#### Implementation Details
```go
type ColumnDefinition struct {
    Key         string                    // Unique identifier
    Title       string                    // Display name
    Description string                    // Help text
    Width       ColumnWidth              // Sizing behavior
    Formatter   func(api.Torrent) string // Data extraction
    Sortable    bool                     // Can be used for sorting
    Default     bool                     // Shown in default view
}

type ColumnWidth struct {
    Min      int     // Minimum width
    Max      int     // Maximum width (0 = unlimited)
    Preferred int    // Preferred width
    Flex     float64 // Growth factor
}
```

#### New Columns to Add
| Column | Key | Description | Default |
|--------|-----|-------------|---------|
| Added On | `added_on` | Date torrent was added | No |
| Completed | `completed_on` | Date download completed | No |
| Last Activity | `last_seen` | Last peer activity | No |
| ETA | `eta` | Estimated time to completion | Yes |
| Category | `category` | Torrent category | No |
| Tags | `tags` | Torrent tags | No |
| Tracker | `tracker` | Primary tracker domain | No |
| Priority | `priority` | Download priority | No |
| Save Path | `save_path` | Download location | No |

#### Configuration UI Flow
```
Press 'C' → Column Configuration
┌─────────────────────────────────────┐
│ Configure Columns                   │
├─────────────────────────────────────┤
│ [x] Name                           │
│ [x] Size                           │  
│ [x] Progress                       │
│ [x] State                          │
│ [ ] Added On                       │
│ [ ] Category                       │
│ [ ] Tags                           │
│                                     │
│ Presets: Basic | Advanced | Custom │
│ [Apply] [Cancel] [Reset]           │
└─────────────────────────────────────┘
```

---

## 🏗️ Implementation Plan

### Phase 1: Responsive Layout ✅ COMPLETED
**Goals**: Full terminal width utilization
- [x] Implement dynamic column sizing system
- [x] Add responsive breakpoints
- [x] Update torrent list component
- [x] Test on various terminal sizes
- [x] Update documentation

### Phase 2: Torrent Sorting ✅ COMPLETED
**Goals**: Multi-column sorting with persistence
- [x] Implement sort configuration types
- [x] Add sorting logic to torrent list
- [x] Create keyboard shortcuts
- [x] Add visual sort indicators
- [x] Persist sort preferences
- [x] Add sorting tests

### Phase 3: Column Configuration ✅ COMPLETED
**Goals**: Configurable columns with UI
- [x] Define all available columns
- [x] Create column definition system
- [x] Implement column configuration UI
- [x] Add configuration persistence (via visible columns tracking)
- [x] Create column toggle functionality
- [x] Comprehensive testing

## 🎨 UX Design Principles

### Responsive Design
- **Mobile-first approach** - Start with 80 columns, expand upward
- **Progressive enhancement** - More features on wider screens
- **Graceful degradation** - Hide non-essential columns when needed

### Sorting
- **Intuitive shortcuts** - Number keys for quick column sorting
- **Visual feedback** - Clear sort direction indicators
- **Sensible defaults** - Sort by name initially, with size as secondary

### Configuration
- **Discoverability** - Make column config easy to find
- **Presets over complexity** - Common configurations built-in
- **Non-destructive** - Easy to reset to defaults

## 🧪 Testing Strategy

### Responsive Layout Testing
- Test on 80, 120, 160, 200+ column terminals
- Verify readability at all sizes
- Test resize behavior during runtime

### Sorting Testing
- Unit tests for all sort algorithms
- Performance testing with 1000+ torrents
- Test sort stability and consistency

### Column Configuration Testing
- Test all column formatters
- Verify configuration persistence
- Test preset loading/saving

## 📋 Success Criteria

### Phase 1 Success ✅
- [x] Utilizes 95%+ of available terminal width
- [x] Maintains readability on 80-column terminals
- [x] Smooth resize behavior

### Phase 2 Success ✅
- [x] Can sort by any visible column
- [x] Sort preferences persist across sessions
- [x] Performance acceptable with 1000+ torrents

### Phase 3 Success ✅
- [x] Users can customize any column visibility
- [x] Configuration UI is intuitive
- [x] All qBittorrent data fields available

## 🔮 Future Considerations

### Advanced Features (Future)
- **Column reordering** - Drag-and-drop column positions
- **Custom column widths** - User-defined pixel/character widths  
- **Saved views** - Multiple named column configurations
- **Export/import configs** - Share column setups
- **Theme-aware sizing** - Different configs per theme

### Performance Optimizations
- **Virtual scrolling** - Handle 10,000+ torrents smoothly
- **Lazy column rendering** - Only render visible columns
- **Sort result caching** - Cache expensive sort operations

---

*This roadmap represents a significant enhancement to qBittorrent TUI's usability and power-user features while maintaining the clean, keyboard-driven interface users expect.*