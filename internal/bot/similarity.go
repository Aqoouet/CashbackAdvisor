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

// findSimilarCategory находит наиболее похожую категорию и возвращает саму категорию,
// процент похожести и расстояние Левенштейна. Мы оцениваем похожесть как по полной
// строке, так и по ключевым словам внутри категории, чтобы лучше ловить опечатки
// в коротких словах («Таксу» -> «Такси»), а также когда в запросе есть служебные
// слова («в городе» и т.п.).
func findSimilarCategory(input string, categories []string) (string, float64, int) {
	if len(categories) == 0 {
		return "", 0, -1
	}

	normalizedInput := strings.ToLower(strings.TrimSpace(input))
	inputWords := filterWords(normalizedInput)
	mainInputWord := ""
	if len(inputWords) > 0 {
		mainInputWord = inputWords[0]
	}

	bestMatch := categories[0]
	bestDistance := levenshteinDistance(normalizedInput, strings.ToLower(categories[0]))
	bestSim := similarity(normalizedInput, strings.ToLower(categories[0]))

	updateBest := func(cat string, dist int, sim float64) {
		if sim > bestSim || (sim == bestSim && dist < bestDistance) {
			bestSim = sim
			bestDistance = dist
			bestMatch = cat
		}
	}

	for _, cat := range categories[1:] {
		catLower := strings.ToLower(cat)

		// Похожесть по полной строке
		distFull := levenshteinDistance(normalizedInput, catLower)
		simFull := similarity(normalizedInput, catLower)
		updateBest(cat, distFull, simFull)

		// Похожесть по ключевым словам категории
		catWords := filterWords(catLower)
		for _, w := range catWords {
			simWord := similarity(normalizedInput, w)
			distWord := levenshteinDistance(normalizedInput, w)
			updateBest(cat, distWord, simWord)

			// Если есть главное слово в запросе, сравним его с словами категории
			if mainInputWord != "" {
				simMain := similarity(mainInputWord, w)
				distMain := levenshteinDistance(mainInputWord, w)
				updateBest(cat, distMain, simMain)
			}
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

