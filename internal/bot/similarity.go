package bot

import (
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

// findSimilarCategory находит наиболее похожую категорию
func findSimilarCategory(input string, categories []string) (string, int) {
	if len(categories) == 0 {
		return "", -1
	}

	bestMatch := categories[0]
	bestDistance := levenshteinDistance(input, categories[0])

	for _, cat := range categories[1:] {
		distance := levenshteinDistance(input, cat)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = cat
		}
	}

	return bestMatch, bestDistance
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

