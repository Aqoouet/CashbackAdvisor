package bot

import (
	"math"
	"strings"
)

// levenshteinDistance вычисляет расстояние Левенштейна между двумя строками
func levenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)
	
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Создаем матрицу
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Инициализация первой строки и столбца
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Вычисление расстояния
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // удаление
				matrix[i][j-1]+1,      // вставка
				matrix[i-1][j-1]+cost, // замена
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min возвращает минимум из трех чисел
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// findSimilarCategory находит наиболее похожую категорию и возвращает саму категорию,
// процент похожести и усреднённое расстояние Левенштейна. Порядок слов не важен:
// для каждого слова запроса берётся лучшее совпадение среди слов категории, затем
// считаем среднее по словам запроса.
func findSimilarCategory(input string, categories []string) (string, float64, int) {
	if len(categories) == 0 {
		return "", 0, -1
	}

	normalizedInput := strings.ToLower(strings.TrimSpace(input))
	inputWords := filterWords(normalizedInput)
	if len(inputWords) == 0 {
		inputWords = []string{normalizedInput}
	}

	bestMatch := categories[0]
	bestSim, bestDistance := scoreCategory(inputWords, categories[0])

	for _, cat := range categories[1:] {
		sim, dist := scoreCategory(inputWords, cat)
		if sim > bestSim || (sim == bestSim && dist < bestDistance) {
			bestSim = sim
			bestDistance = dist
			bestMatch = cat
		}
	}

	return bestMatch, bestSim, bestDistance
}

// similarity вычисляет коэффициент похожести (0-100)
func similarity(s1, s2 string) float64 {
	maxLen := float64(max(len(s1), len(s2)))
	if maxLen == 0 {
		return 100.0
	}
	
	distance := float64(levenshteinDistance(s1, s2))
	return (1.0 - distance/maxLen) * 100.0
}

// max возвращает максимум из двух чисел
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// filterWords разбивает строку на слова, убирая короткие служебные слова
func filterWords(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') &&
			!(r >= 'A' && r <= 'Z') &&
			!(r >= 'а' && r <= 'я') &&
			!(r >= 'А' && r <= 'Я') &&
			r != 'ё' && r != 'Ё'
	})

	stop := map[string]bool{
		"в": true, "во": true, "на": true, "и": true, "для": true, "из": true,
		"к": true, "по": true, "с": true, "со": true, "от": true, "до": true,
		"у": true, "за": true, "над": true, "под": true, "при": true,
		"город": true, "городе": true, "городом": true,
	}

	var res []string
	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		if stop[p] {
			continue
		}
		res = append(res, p)
	}
	return res
}

// scoreCategory считает похожесть категории: порядок слов не важен, берём
// для каждого слова запроса лучшее совпадение по словам категории и усредняем.
// Улучшенная версия с бонусами за точные совпадения и вхождения.
func scoreCategory(inputWords []string, category string) (float64, int) {
	catLower := strings.ToLower(category)
	catWords := filterWords(catLower)
	if len(catWords) == 0 {
		catWords = []string{catLower}
	}

	totalSim := 0.0
	totalDist := 0
	exactMatches := 0

	for _, iw := range inputWords {
		bestSim := 0.0
		bestDist := len(iw) + len(catLower) // заведомо худшее
		hasExactMatch := false
		
		// Проверяем точное вхождение подстроки в категорию
		if strings.Contains(catLower, strings.ToLower(iw)) {
			// Бонус за точное вхождение
			bestSim = 95.0
			bestDist = 0
			hasExactMatch = true
		} else {
			// Ищем лучшее совпадение среди слов категории
			for _, cw := range catWords {
				// Проверяем точное совпадение слов
				if strings.ToLower(iw) == strings.ToLower(cw) {
					bestSim = 100.0
					bestDist = 0
					hasExactMatch = true
					break
				}
				
				// Проверяем начинается ли слово категории с слова запроса
				if strings.HasPrefix(strings.ToLower(cw), strings.ToLower(iw)) {
					s := 90.0 // Бонус за совпадение префикса
					d := len(cw) - len(iw)
					if s > bestSim || (s == bestSim && d < bestDist) {
						bestSim = s
						bestDist = d
					}
				} else {
					// Обычное сравнение по Левенштейну
					s := similarity(iw, cw)
					d := levenshteinDistance(iw, cw)
					if s > bestSim || (s == bestSim && d < bestDist) {
						bestSim = s
						bestDist = d
					}
				}
			}
		}
		
		if hasExactMatch {
			exactMatches++
		}
		
		totalSim += bestSim
		totalDist += bestDist
	}

	avgSim := totalSim / float64(len(inputWords))
	avgDist := int(math.Round(float64(totalDist) / float64(len(inputWords))))
	
	// Бонус, если все слова запроса точно совпали
	if exactMatches == len(inputWords) && len(inputWords) > 0 {
		avgSim = math.Min(100.0, avgSim+5.0)
	}

	return avgSim, avgDist
}

