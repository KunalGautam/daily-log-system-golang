package analytics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MoodEntry struct {
	Date       string `json:"date"`
	Score      int    `json:"score"`
	Activities string `json:"activities"`
}

type MoodTrends struct {
	Labels    []string    `json:"labels"`
	Scores    []float64   `json:"scores"`
	Average   float64     `json:"average"`
	Stability float64     `json:"stability"`
	Entries   []MoodEntry `json:"entries"`
}

type ActivityCount struct {
	Activity string `json:"activity"`
	Count    int64  `json:"count"`
}

type ActivityCorrelation struct {
	Activity1   string  `json:"activity1"`
	Activity2   string  `json:"activity2"`
	Correlation float64 `json:"correlation"`
	Count       int     `json:"count"`
}

type HabitSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CompletionRate float64   `json:"completion_rate"`
	Streak         int       `json:"streak"`
}

type HabitAnalytics struct {
	TotalHabits        int            `json:"total_habits"`
	ActiveHabits       int            `json:"active_habits"`
	AvgStreak          float64        `json:"avg_streak"`
	AvgCompletionRate  float64        `json:"avg_completion_rate"`
	TopHabits          []HabitSummary `json:"top_habits"`
}

type GoalAnalytics struct {
	TotalGoals    int     `json:"total_goals"`
	CompletedGoals int    `json:"completed_goals"`
	InProgress    int     `json:"in_progress"`
	AvgProgress   float64 `json:"avg_progress"`
}

type DayCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type DayMood struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
}

type TimeAnalytics struct {
	Period        string          `json:"period"`
	EntriesByDay  []DayCount      `json:"entries_by_day"`
	MoodByDay     []DayMood       `json:"mood_by_day"`
	TopActivities []ActivityCount `json:"top_activities"`
	TotalEntries  int             `json:"total_entries"`
}

type streakInfo struct {
	HabitID   uuid.UUID `json:"habit_id"`
	HabitName string    `json:"habit_name"`
	Streak    int       `json:"streak"`
}

type RecentEntry struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	EntryType string    `json:"entry_type"`
	MoodScore *int      `json:"mood_score"`
	CreatedAt time.Time `json:"created_at"`
}

type GoalInfo struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	TargetValue     float64   `json:"target_value"`
	CurrentValue    float64   `json:"current_value"`
	ProgressPercent float64   `json:"progress_percent"`
}

type DashboardSummary struct {
	TodayMood      *int          `json:"today_mood"`
	EntryCount     int64         `json:"entry_count"`
	WeekEntryCount int64         `json:"week_entry_count"`
	CurrentStreaks []streakInfo  `json:"current_streaks"`
	RecentEntries  []RecentEntry `json:"recent_entries"`
	UpcomingGoals  []GoalInfo    `json:"upcoming_goals"`
	ActiveHabits   int64         `json:"active_habits"`
}

type SystemStats struct {
	TotalUsers      int64  `json:"total_users"`
	TotalEntries    int64  `json:"total_entries"`
	TotalHabits     int64  `json:"total_habits"`
	TotalGoals      int64  `json:"total_goals"`
	ActiveUsers24h  int64  `json:"active_users_24h"`
	StorageUsed     string `json:"storage_used"`
	DBConnections   int    `json:"db_connections"`
}

type AnalyticsService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (s *AnalyticsService) GetMoodTrends(userID uuid.UUID, period string, startDate, endDate time.Time) (*MoodTrends, error) {
	var dateTrunc string
	switch period {
	case "weekly":
		dateTrunc = "DATE_TRUNC('week', entry_date)::date"
	case "monthly":
		dateTrunc = "DATE_TRUNC('month', entry_date)::date"
	default:
		dateTrunc = "DATE_TRUNC('day', entry_date)::date"
	}

	type moodRow struct {
		Label      string
		AvgScore   float64
		EntryCount int
	}
	var rows []moodRow
	if err := s.db.Table("entries").
		Select(fmt.Sprintf("%s as label, AVG(mood_score) as avg_score, COUNT(*) as entry_count", dateTrunc)).
		Where("user_id = ? AND mood_score IS NOT NULL AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Group("label").
		Order("label ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	var entries []MoodEntry
	var allScores []float64
	result := &MoodTrends{}

	for _, r := range rows {
		result.Labels = append(result.Labels, r.Label)
		result.Scores = append(result.Scores, r.AvgScore)
		for i := 0; i < r.EntryCount; i++ {
			allScores = append(allScores, r.AvgScore)
		}
	}

	if err := s.db.Table("entries").
		Where("user_id = ? AND mood_score IS NOT NULL AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Order("entry_date ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	result.Entries = entries

	var total float64
	var count int
	for _, e := range entries {
		total += float64(e.Score)
		count++
	}
	if count > 0 {
		result.Average = math.Round(total/float64(count)*100) / 100
	}

	if len(allScores) > 0 {
		mean := total / float64(count)
		var variance float64
		for _, s := range allScores {
			diff := s - mean
			variance += diff * diff
		}
		variance /= float64(len(allScores))
		result.Stability = math.Round((1-math.Sqrt(variance)/5)*100) / 100
		if result.Stability < 0 {
			result.Stability = 0
		}
	}

	return result, nil
}

func (s *AnalyticsService) GetMoodDistribution(userID uuid.UUID, startDate, endDate time.Time) (map[int]int64, error) {
	type moodCount struct {
		MoodScore int
		Count     int64
	}
	var rows []moodCount
	if err := s.db.Table("entries").
		Select("mood_score, COUNT(*) as count").
		Where("user_id = ? AND mood_score IS NOT NULL AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Group("mood_score").
		Order("mood_score ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[int]int64)
	for _, r := range rows {
		result[r.MoodScore] = r.Count
	}
	return result, nil
}

func (s *AnalyticsService) GetActivityFrequency(userID uuid.UUID, startDate, endDate time.Time) ([]ActivityCount, error) {
	type row struct {
		Activity string
		Count    int64
	}
	var rows []row
	if err := s.db.Table("entries").
		Select("unnest(string_to_array(activities, ',')) as activity, COUNT(*) as count").
		Where("user_id = ? AND activities IS NOT NULL AND activities != '' AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Group("activity").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ActivityCount, len(rows))
	for i, r := range rows {
		result[i] = ActivityCount{
			Activity: strings.TrimSpace(r.Activity),
			Count:    r.Count,
		}
	}
	return result, nil
}

func (s *AnalyticsService) GetActivityCorrelations(userID uuid.UUID, startDate, endDate time.Time) ([]ActivityCorrelation, error) {
	type entryActivities struct {
		Activities string
	}
	var entries []entryActivities
	if err := s.db.Table("entries").
		Select("activities").
		Where("user_id = ? AND activities IS NOT NULL AND activities != '' AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	activitySets := make([][]string, 0, len(entries))
	allActivities := make(map[string]bool)
	for _, e := range entries {
		parts := strings.Split(e.Activities, ",")
		set := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				allActivities[p] = true
				set = append(set, p)
			}
		}
		if len(set) > 0 {
			activitySets = append(activitySets, set)
		}
	}

	activityList := make([]string, 0, len(allActivities))
	for a := range allActivities {
		activityList = append(activityList, a)
	}
	sort.Strings(activityList)

	totalEntries := len(activitySets)

	cooccurrence := make(map[string]map[string]int)
	for _, a := range activityList {
		cooccurrence[a] = make(map[string]int)
	}

	for _, set := range activitySets {
		for i := 0; i < len(set); i++ {
			for j := i + 1; j < len(set); j++ {
				a1, a2 := set[i], set[j]
				if a1 > a2 {
					a1, a2 = a2, a1
				}
				cooccurrence[a1][a2]++
			}
		}
	}

	singleCounts := make(map[string]int)
	for _, set := range activitySets {
		for _, a := range set {
			singleCounts[a]++
		}
	}

	var correlations []ActivityCorrelation

	for _, a1 := range activityList {
		for _, a2 := range activityList {
			if a1 >= a2 {
				continue
			}
			co := cooccurrence[a1][a2]
			if co < 2 {
				continue
			}

			n1 := singleCounts[a1]
			n2 := singleCounts[a2]

			expected := float64(n1) * float64(n2) / float64(totalEntries)
			var corr float64
			if expected > 0 {
				corr = (float64(co) - expected) / expected
			} else {
				corr = 0
			}
			corr = math.Round(corr*100) / 100

			correlations = append(correlations, ActivityCorrelation{
				Activity1:   a1,
				Activity2:   a2,
				Correlation: corr,
				Count:       co,
			})
		}
	}

	sort.Slice(correlations, func(i, j int) bool {
		return correlations[i].Correlation > correlations[j].Correlation
	})

	if len(correlations) > 50 {
		correlations = correlations[:50]
	}

	return correlations, nil
}

func (s *AnalyticsService) GetHabitAnalytics(userID uuid.UUID) (*HabitAnalytics, error) {
	type habitRow struct {
		ID              uuid.UUID
		Name            string
		IsArchived      bool
		CurrentStreak   int
		TotalCompletions int
		SuccessRate     float64
	}
	var habits []habitRow
	if err := s.db.Table("habits").
		Select("id, name, is_archived, current_streak, total_completions, success_rate").
		Where("user_id = ?", userID).
		Find(&habits).Error; err != nil {
		return nil, err
	}

	result := &HabitAnalytics{}
	var totalStreak float64
	var habitCount int

	for _, h := range habits {
		result.TotalHabits++
		if !h.IsArchived {
			result.ActiveHabits++
		}
		totalStreak += float64(h.CurrentStreak)
		habitCount++
	}

	if habitCount > 0 {
		result.AvgStreak = math.Round(totalStreak/float64(habitCount)*100) / 100
	}

	var totalCompletionRate float64
	var completionCount int
	topHabits := make([]HabitSummary, 0, len(habits))

	for _, h := range habits {
		rate := h.SuccessRate
		if rate == 0 && h.TotalCompletions > 0 {
			rate = float64(h.TotalCompletions) * 100
		}
		totalCompletionRate += rate
		completionCount++

		topHabits = append(topHabits, HabitSummary{
			ID:             h.ID,
			Name:           h.Name,
			CompletionRate: rate,
			Streak:         h.CurrentStreak,
		})
	}

	if completionCount > 0 {
		result.AvgCompletionRate = math.Round(totalCompletionRate/float64(completionCount)*100) / 100
	}

	sort.Slice(topHabits, func(i, j int) bool {
		return topHabits[i].CompletionRate > topHabits[j].CompletionRate
	})
	if len(topHabits) > 5 {
		topHabits = topHabits[:5]
	}
	result.TopHabits = topHabits

	return result, nil
}

func (s *AnalyticsService) GetGoalAnalytics(userID uuid.UUID) (*GoalAnalytics, error) {
	type goalRow struct {
		IsCompleted  bool
		CurrentValue float64
		TargetValue  float64
	}
	var goals []goalRow
	if err := s.db.Table("goals").
		Select("is_completed, current_value, target_value").
		Where("user_id = ?", userID).
		Find(&goals).Error; err != nil {
		return nil, err
	}

	result := &GoalAnalytics{}
	var totalProgress float64
	var progressCount int

	for _, g := range goals {
		result.TotalGoals++
		if g.IsCompleted {
			result.CompletedGoals++
		} else {
			result.InProgress++
		}
		if g.TargetValue > 0 {
			totalProgress += (g.CurrentValue / g.TargetValue) * 100
			progressCount++
		}
	}

	if progressCount > 0 {
		result.AvgProgress = math.Round(totalProgress/float64(progressCount)*100) / 100
	}

	return result, nil
}

func (s *AnalyticsService) GetTimeAnalytics(userID uuid.UUID, period string, year, month int) (*TimeAnalytics, error) {
	var startDate, endDate time.Time
	loc := time.UTC

	switch period {
	case "monthly":
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1)
	case "yearly":
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(1, 0, -1)
	default:
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1)
	}

	var totalEntries int64
	s.db.Table("entries").Where("user_id = ? AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).Count(&totalEntries)

	type dayCountRow struct {
		Date  string
		Count int64
	}
	var dayRows []dayCountRow
	s.db.Table("entries").
		Select("TO_CHAR(entry_date, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("user_id = ? AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Group("date").
		Order("date ASC").
		Scan(&dayRows)

	entriesByDay := make([]DayCount, len(dayRows))
	for i, r := range dayRows {
		entriesByDay[i] = DayCount{Date: r.Date, Count: r.Count}
	}

	type dayMoodRow struct {
		Date    string
		AvgMood float64
	}
	var moodRows []dayMoodRow
	s.db.Table("entries").
		Select("TO_CHAR(entry_date, 'YYYY-MM-DD') as date, AVG(mood_score) as avg_mood").
		Where("user_id = ? AND mood_score IS NOT NULL AND entry_date >= ? AND entry_date <= ?", userID, startDate, endDate).
		Group("date").
		Order("date ASC").
		Scan(&moodRows)

	moodByDay := make([]DayMood, len(moodRows))
	for i, r := range moodRows {
		moodByDay[i] = DayMood{Date: r.Date, Score: math.Round(r.AvgMood*100) / 100}
	}

	topActivities, _ := s.GetActivityFrequency(userID, startDate, endDate)
	if len(topActivities) > 10 {
		topActivities = topActivities[:10]
	}

	return &TimeAnalytics{
		Period:        period,
		EntriesByDay:  entriesByDay,
		MoodByDay:     moodByDay,
		TopActivities: topActivities,
		TotalEntries:  int(totalEntries),
	}, nil
}

func (s *AnalyticsService) GetDashboardSummary(userID uuid.UUID) (*DashboardSummary, error) {
	result := &DashboardSummary{}

	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)

	var todayEntry struct {
		MoodScore *int
	}
	s.db.Table("entries").
		Select("mood_score").
		Where("user_id = ? AND entry_date = ? AND mood_score IS NOT NULL", userID, today).
		First(&todayEntry)
	result.TodayMood = todayEntry.MoodScore

	s.db.Table("entries").Where("user_id = ?", userID).Count(&result.EntryCount)

	s.db.Table("entries").Where("user_id = ? AND entry_date >= ?", userID, weekAgo).Count(&result.WeekEntryCount)

	type streakRow struct {
		ID            uuid.UUID
		Name          string
		CurrentStreak int
	}
	var streaks []streakRow
	s.db.Table("habits").
		Select("id, name, current_streak").
		Where("user_id = ? AND is_archived = ? AND current_streak > ?", userID, false, 0).
		Order("current_streak DESC").
		Find(&streaks)

	result.CurrentStreaks = make([]streakInfo, len(streaks))
	for i, s := range streaks {
		result.CurrentStreaks[i] = streakInfo{
			HabitID:   s.ID,
			HabitName: s.Name,
			Streak:    s.CurrentStreak,
		}
	}

	type recentRow struct {
		ID        uuid.UUID
		Title     *string
		EntryType string
		MoodScore *int
		CreatedAt time.Time
	}
	var recent []recentRow
	s.db.Table("entries").
		Select("id, title, entry_type, mood_score, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&recent)

	result.RecentEntries = make([]RecentEntry, len(recent))
	for i, r := range recent {
		result.RecentEntries[i] = RecentEntry{
			ID:        r.ID,
			Title:     r.Title,
			EntryType: r.EntryType,
			MoodScore: r.MoodScore,
			CreatedAt: r.CreatedAt,
		}
	}

	type goalRow struct {
		ID           uuid.UUID
		Title        string
		TargetValue  float64
		CurrentValue float64
	}
	var goals []goalRow
	s.db.Table("goals").
		Select("id, title, target_value, current_value").
		Where("user_id = ? AND is_completed = ? AND is_archived = ?", userID, false, false).
		Order("(current_value / NULLIF(target_value, 0)) DESC").
		Limit(5).
		Find(&goals)

	result.UpcomingGoals = make([]GoalInfo, len(goals))
	for i, g := range goals {
		var progress float64
		if g.TargetValue > 0 {
			progress = math.Round((g.CurrentValue/g.TargetValue)*10000) / 100
		}
		result.UpcomingGoals[i] = GoalInfo{
			ID:              g.ID,
			Title:           g.Title,
			TargetValue:     g.TargetValue,
			CurrentValue:    g.CurrentValue,
			ProgressPercent: progress,
		}
	}

	s.db.Table("habits").
		Where("user_id = ? AND is_archived = ?", userID, false).
		Count(&result.ActiveHabits)

	return result, nil
}

func (s *AnalyticsService) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	s.db.Table("users").Count(&stats.TotalUsers)
	s.db.Table("entries").Count(&stats.TotalEntries)
	s.db.Table("habits").Count(&stats.TotalHabits)
	s.db.Table("goals").Count(&stats.TotalGoals)

	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	s.db.Table("entries").
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", twentyFourHoursAgo).
		Scan(&stats.ActiveUsers24h)

	sqlDB, err := s.db.DB()
	if err == nil {
		stats.DBConnections = sqlDB.Stats().OpenConnections
	}

	stats.StorageUsed = "0 B"

	return stats, nil
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	uid, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	uidStr, ok := uid.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID is not a string")
	}
	return uuid.Parse(uidStr)
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}
	endDate = endDate.Add(24*time.Hour - time.Nanosecond)

	return startDate, endDate, nil
}

func (s *AnalyticsService) HandleMoodTrends(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	period := c.DefaultQuery("period", "daily")
	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trends, err := s.GetMoodTrends(userID, period, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

func (s *AnalyticsService) HandleMoodDistribution(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	distribution, err := s.GetMoodDistribution(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, distribution)
}

func (s *AnalyticsService) HandleActivityFrequency(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	freq, err := s.GetActivityFrequency(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, freq)
}

func (s *AnalyticsService) HandleActivityCorrelations(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	correlations, err := s.GetActivityCorrelations(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, correlations)
}

func (s *AnalyticsService) HandleHabitAnalytics(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	analytics, err := s.GetHabitAnalytics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (s *AnalyticsService) HandleGoalAnalytics(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	analytics, err := s.GetGoalAnalytics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (s *AnalyticsService) HandleTimeAnalytics(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	period := c.DefaultQuery("period", "monthly")
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))

	analytics, err := s.GetTimeAnalytics(userID, period, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (s *AnalyticsService) HandleDashboard(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	summary, err := s.GetDashboardSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (s *AnalyticsService) HandleSystemStats(c *gin.Context) {
	stats, err := s.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func RegisterRoutes(rg *gin.RouterGroup, analyticsSvc *AnalyticsService, mw gin.HandlerFunc) {
	analytics := rg.Group("/analytics").Use(mw)
	{
		analytics.GET("/dashboard", analyticsSvc.HandleDashboard)
		analytics.GET("/mood/trends", analyticsSvc.HandleMoodTrends)
		analytics.GET("/mood/distribution", analyticsSvc.HandleMoodDistribution)
		analytics.GET("/activities/frequency", analyticsSvc.HandleActivityFrequency)
		analytics.GET("/activities/correlations", analyticsSvc.HandleActivityCorrelations)
		analytics.GET("/habits", analyticsSvc.HandleHabitAnalytics)
		analytics.GET("/goals", analyticsSvc.HandleGoalAnalytics)
		analytics.GET("/time", analyticsSvc.HandleTimeAnalytics)
	}
}
