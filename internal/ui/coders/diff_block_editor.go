package coders

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/coding-hui/common/util/fileutil"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

const (
	HEAD    = "<<<<<<< SEARCH"
	DIVIDER = "======="
	UPDATED = ">>>>>>> REPLACE"

	similarityThreshold = 0.8
	similarityScale     = 0.1
)

var (
	separators = regexp.QuoteMeta(HEAD) + "|" + regexp.QuoteMeta(DIVIDER) + "|" + regexp.QuoteMeta(UPDATED)
)

type pathAndCode struct {
	path, code string
}

type EditBlockCoder struct {
	coder                  *AutoCoder
	fence                  []string
	partialResponseContent string
}

func NewEditBlockCoder(coder *AutoCoder, fence []string) *EditBlockCoder {
	return &EditBlockCoder{
		coder: coder,
		fence: fence,
	}
}

func (e *EditBlockCoder) Name() string {
	return "edit_block_coder"
}

func (e *EditBlockCoder) Prompt() prompts.ChatPromptTemplate {
	return promptBaseCoder
}

func (e *EditBlockCoder) FormatMessages(values map[string]any) ([]llms.MessageContent, error) {
	return formatPrompt(e.Prompt(), values)
}

func (e *EditBlockCoder) GetEdits(_ context.Context) ([]PartialCodeBlock, error) {
	openFence, closeFence := chooseBestFence(e.partialResponseContent)
	if len(e.fence) == 2 {
		openFence, closeFence = e.fence[0], e.fence[1]
	}

	edits, err := findOriginalUpdateBlocks(e.partialResponseContent, []string{openFence, closeFence})
	if err != nil {
		return nil, err
	}

	return edits, nil
}

func (e *EditBlockCoder) GetModifiedFiles(ctx context.Context) ([]string, error) {
	edits, err := e.GetEdits(ctx)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(edits))
	for _, edit := range edits {
		files = append(files, edit.Path)
	}

	return files, nil
}

func (e *EditBlockCoder) UpdateCodeFences(_ context.Context, code string) (string, string) {
	e.fence = make([]string, 2)
	e.fence[0], e.fence[1] = chooseBestFence(code)

	return e.fence[0], e.fence[1]
}

func (e *EditBlockCoder) ApplyEdits(ctx context.Context, edits []PartialCodeBlock, needConfirm bool) error {
	var failed []PartialCodeBlock

	for _, block := range edits {
		if err := e.applyEdit(ctx, block, needConfirm); err != nil {
			failed = append(failed, block)
		}
	}

	if len(failed) > 0 {
		return e.handleFailedEdits(failed)
	}

	return nil
}

func (e *EditBlockCoder) applyEdit(_ context.Context, block PartialCodeBlock, needConfirm bool) error {
	absPath, err := absFilePath(e.coder.codeBasePath, block.Path)
	if err != nil {
		return err
	}

	fileExists, err := fileutil.FileExists(absPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if !fileExists {
		if ok := console.WaitForUserConfirm("Whether to create the %s file? (Y/n)", block.Path); ok {
			if err := fileutil.WriteFile(absPath, []byte("")); err != nil {
				return err
			}
			console.Render("Created %s file", block.Path)
		} else {
			return errbook.NewUserErrorf("Apply %s edit cancelled, file cannot be found", block.Path)
		}
	}

	rawFileContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	fileExt := strings.TrimLeft(filepath.Ext(absPath), ".")
	confirmMsg := fmt.Sprintf(`
%s
%s%s
<<<<<<< SEARCH
%s
=======
%s
>>>>>>> REPLACE
%s

Are you sure you want to apply these edits? (Y/n)`,
		block.Path,
		e.fence[0],
		fileExt,
		block.OriginalText,
		block.UpdatedText,
		e.fence[1],
	)

	if needConfirm {
		if ok := console.WaitForUserConfirmApplyChanges(confirmMsg); !ok {
			return errbook.NewUserErrorf("Apply %s edit cancelled", block.Path)
		}
	}

	newFileContent := doReplace(absPath, string(rawFileContent), block.OriginalText, block.UpdatedText, e.fence)
	if len(newFileContent) == 0 {
		return errbook.New("Code block is empty and cannot be updated to file %s", block.Path)
	}

	err = fileutil.WriteFile(absPath, []byte(newFileContent))
	if err != nil {
		return err
	}

	console.Render("Applied %s edit", block.Path)

	return nil
}

func (e *EditBlockCoder) handleFailedEdits(failed []PartialCodeBlock) error {
	blocks := "block"
	if len(failed) > 1 {
		blocks = "blocks"
	}

	errMsg := fmt.Sprintf("# %d SEARCH/REPLACE %s failed to match!\n", len(failed), blocks)

	for _, block := range failed {
		absPath, err := absFilePath(e.coder.codeBasePath, block.Path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}

		errMsg += fmt.Sprintf(`
## SearchReplaceNoExactMatch: This SEARCH block failed to exactly match lines in %s
<<<<<<< SEARCH
%s=======
%s>>>>>>> REPLACE

`, block.Path, block.OriginalText, block.UpdatedText)

		didYouMean := findSimilarLines(block.OriginalText, string(content))
		if len(didYouMean) > 0 {
			errMsg += fmt.Sprintf(`Did you mean to match some of these actual lines from %s?

%s
%s
%s

`, block.Path, e.fence[0], didYouMean, e.fence[1])
		}

		if strings.Contains(string(content), block.UpdatedText) {
			errMsg += fmt.Sprintf(`Are you sure you need this SEARCH/REPLACE block?
The REPLACE lines are already in %s!

The SEARCH section must exactly match an existing block of lines including all white  space, comments, indentation, docstrings, etc.
`, block.Path)
		}
		console.Render(errMsg)
	}

	return nil
}

func (e *EditBlockCoder) Execute(ctx context.Context, messages []llms.MessageContent) error {
	//e.coder.Loading("Please wait while we design the code")

	go func() {
		for output := range e.coder.engine.GetChannel() {
			if output.IsLast() {
				fmt.Println()
			}
			fmt.Print(output.GetContent())
		}
	}()

	output, err := e.coder.engine.ChatStream(ctx, messages)
	if err != nil {
		return err
	}

	e.partialResponseContent = output.Choices[0].Content

	if ok := console.WaitForUserConfirmApplyChanges("%s\n\nAre you sure you want to apply these codes? (Y/n)", e.partialResponseContent); !ok {
		return errbook.NewUserErrorf("Apply %s edit cancelled", e.partialResponseContent)
	}

	edits, err := e.GetEdits(ctx)
	if err != nil {
	}

	if len(edits) <= 0 {
		return errbook.New("No edits were made")
	}

	err = e.ApplyEdits(ctx, edits, false)
	if err != nil {
		return err
	}

	return nil
}

func findOriginalUpdateBlocks(content string, fence []string) ([]PartialCodeBlock, error) {
	edits := findAllCodeBlocks(content, fence)
	result := make([]PartialCodeBlock, 0, len(edits))

	for _, edit := range edits {
		block, err := parseEditBlock(edit)
		if err != nil {
			return nil, err
		}
		result = append(result, block)
	}

	return result, nil
}

func parseEditBlock(edit pathAndCode) (PartialCodeBlock, error) {
	heads := []string{}
	updates := []string{}
	inHead := false
	inUpdated := false

	for _, line := range strings.Split(edit.code, "\n") {
		if strings.TrimSpace(line) == HEAD {
			inHead = true
			continue
		}

		if strings.TrimSpace(line) == DIVIDER {
			inHead = false
			inUpdated = true
			continue
		}

		if strings.TrimSpace(line) == UPDATED {
			inHead = false
			inUpdated = false
			continue
		}

		if inHead {
			heads = append(heads, line)
		}

		if inUpdated {
			updates = append(updates, line)
		}
	}

	if len(updates) == 0 && len(heads) == 0 {
		return PartialCodeBlock{}, fmt.Errorf("no code block found in %s", edit.path)
	}

	return PartialCodeBlock{
		edit.path,
		strings.Join(heads, "\n"),
		strings.Join(updates, "\n"),
	}, nil
}

func findAllCodeBlocks(content string, fence []string) []pathAndCode {
	lines := strings.Split(content, "\n")

	startMarker := func(line string, idx int) bool {
		return strings.HasPrefix(strings.TrimSpace(line), fence[0]) && idx+1 < len(lines) && strings.TrimSpace(lines[idx+1]) != HEAD
	}

	endMarker := func(line string, idx int) bool {
		return strings.HasPrefix(strings.TrimSpace(line), fence[1]) && strings.TrimSpace(lines[idx-1]) == UPDATED
	}

	blocks := []pathAndCode{}
	currentBlock := []string{}
	startMarkerCount := 0
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if startMarker(line, i) && startMarkerCount == 0 {
			startMarkerCount++
		} else if endMarker(line, i) && startMarkerCount == 1 {
			startMarkerCount--
			if len(currentBlock) > 0 {
				path := strings.TrimSpace(currentBlock[0])
				code := strings.Join(currentBlock[1:], "\n")
				currentBlock = []string{}
				blocks = append(blocks, pathAndCode{path, code})
			}
		} else if startMarkerCount > 0 {
			currentBlock = append(currentBlock, line)
		}
	}

	return blocks
}

func findFilename(line string, fence []string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimRight(line, ":")
	line = strings.TrimLeft(line, "#")
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "`")
	line = strings.Trim(line, "*")
	line = strings.Replace(line, "\\_", "_", -1)

	parts := strings.Split(line, "\n")

	if len(parts) < 2 {
		return ""
	}

	startFence := fence[0]
	if strings.HasPrefix(parts[len(parts)-2], startFence) {
		return parts[len(parts)-1]
	}

	return ""
}

func stripQuotedWrapping(res, fname string, fence []string) string {
	if res == "" {
		return res
	}

	lines := strings.Split(res, "\n")

	if fname != "" && strings.HasSuffix(strings.TrimSpace(lines[0]), filepath.Base(fname)) {
		lines = lines[1:]
	}

	if strings.HasPrefix(lines[0], fence[0]) && strings.HasPrefix(lines[len(lines)-1], fence[1]) {
		lines = lines[1 : len(lines)-1]
	}

	res = strings.Join(lines, "\n")
	if res != "" && res[len(res)-1] != '\n' {
		res += "\n"
	}

	return res
}

func doReplace(fileName string, content, beforeText, afterText string, fence []string) string {
	beforeText = stripQuotedWrapping(beforeText, fileName, fence)
	afterText = stripQuotedWrapping(afterText, fileName, fence)

	if _, err := os.Stat(fileName); os.IsNotExist(err) && beforeText == "" {
		// touch empty file
		err = os.WriteFile(fileName, []byte{}, 0644)
		if err != nil {
			return ""
		}
	}

	if content == "" || beforeText == "" {
		return content + afterText
	}

	newContent := replaceMostSimilarChunk(content, beforeText, afterText)

	return newContent
}

func split(content string) (string, []string) {
	if len(content) >= 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	lines := strings.SplitAfter(content, "\n")
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return content, lines
}

func replaceMostSimilarChunk(rawWhole, rawPart, rawReplace string) string {
	_, wholeLines := split(rawWhole)
	part, partLines := split(rawPart)
	_, replaceLines := split(rawReplace)

	res := perfectOrWhitespace(wholeLines, partLines, replaceLines)
	if len(res) > 0 {
		return res
	}

	// try fuzzy matching
	res = replaceClosestEditDistance(wholeLines, part, partLines, replaceLines)
	if len(res) > 0 {
		return res
	}

	return ""
}

func perfectOrWhitespace(wholeLines, partLines, replaceLines []string) string {
	res := perfectReplace(wholeLines, partLines, replaceLines)
	if len(res) > 0 {
		return res
	}

	res = replacePartWithMissingLeadingWhitespace(wholeLines, partLines, replaceLines)
	if len(res) > 0 {
		return res
	}

	return ""
}

func perfectReplace(wholeLines, partLines, replaceLines []string) string {
	part := strings.Join(partLines, "")
	partLen := len(partLines)

	for i := 0; i < len(wholeLines)-partLen+1; i++ {
		if strings.EqualFold(strings.Join(wholeLines[i:i+partLen], ""), part) {
			// Create a new slice for the result instead of modifying wholeLines
			res := make([]string, 0, len(wholeLines)-partLen+len(replaceLines))
			res = append(res, wholeLines[:i]...)
			res = append(res, replaceLines...)
			res = append(res, wholeLines[i+partLen:]...)
			return strings.Join(res, "")
		}
	}

	return ""
}

// replacePartWithMissingLeadingWhitespace 替换具有缺失前导空白的部分。
func replacePartWithMissingLeadingWhitespace(wholeLines []string, partLines []string, replaceLines []string) string {
	var leading []int

	for _, line := range partLines {
		if strings.TrimSpace(line) != "" {
			leading = append(leading, len(line)-len(strings.TrimLeftFunc(line, unicode.IsSpace)))
		}
	}
	for _, line := range replaceLines {
		if strings.TrimSpace(line) != "" {
			leading = append(leading, len(line)-len(strings.TrimLeftFunc(line, unicode.IsSpace)))
		}
	}

	if len(leading) > 0 {
		minLeading := leading[0]
		for _, l := range leading[1:] {
			if l < minLeading {
				minLeading = l
			}
		}

		// 根据最小的起始空白字符数调整 partLines 和 replaceLines
		partLines = trimLeading(partLines, minLeading)
		replaceLines = trimLeading(replaceLines, minLeading)
	}

	numPartLines := len(partLines)

	for i := 0; i <= len(wholeLines)-numPartLines+1; i++ {
		addLeading := matchButForLeadingWhitespace(wholeLines[i:i+numPartLines], partLines)
		if addLeading == "" {
			continue
		}

		replaceLines = addLeadingWhitespace(replaceLines, addLeading)

		// Create a new slice for the result instead of modifying wholeLines
		result := make([]string, 0, len(wholeLines)-numPartLines+len(replaceLines))
		result = append(result, wholeLines[:i]...)
		result = append(result, replaceLines...)
		result = append(result, wholeLines[i+numPartLines:]...)

		return strings.Join(result, "")
	}

	return ""
}

// trimLeading 从每行字符串中去除指定数量的前导空白。
func trimLeading(lines []string, num int) []string {
	var result []string
	for _, line := range lines {
		if (len(strings.TrimSpace(line))) > 0 {
			result = append(result, strings.TrimLeft(line, strings.Repeat(" ", num)))
		}
	}
	return result
}

// addLeadingWhitespace 向每行字符串添加前导空白。
func addLeadingWhitespace(lines []string, leading string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			result[i] = leading + line
		} else {
			result[i] = line
		}
	}
	return result
}

// matchButForLeadingWhitespace 检查两个字符串切片是否除了前导空白外相同。
func matchButForLeadingWhitespace(wholeLines []string, partLines []string) string {
	allMatch := true
	for i := range wholeLines {
		if strings.TrimSpace(wholeLines[i]) != strings.TrimSpace(partLines[i]) {
			allMatch = false
			break
		}
	}
	if !allMatch {
		return ""
	}

	add := make(map[string]bool)
	for i, line := range wholeLines {
		if strings.TrimSpace(line) != "" && len(line) >= len(partLines) {
			add[line[:len(line)-len(partLines[i])]] = true
		}
	}

	if len(add) != 1 {
		return ""
	}

	for key := range add {
		return key
	}

	return ""
}

func replaceClosestEditDistance(wholeLines []string, part string, partLines []string, replaceLines []string) string {
	var (
		maxSimilarity         = 0.0
		mostSimilarChunkStart = -1
		mostSimilarChunkEnd   = -1

		minLen = int(math.Floor(float64(len(partLines)) * (1 - similarityScale)))
		maxLen = int(math.Ceil(float64(len(partLines)) * (1 + similarityScale)))
	)

	for length := minLen; length <= maxLen; length++ {
		for i := 0; i <= len(wholeLines)-length; i++ {
			chunk := strings.Join(wholeLines[i:i+length], "")

			// Calculate similarity using Levenshtein distance
			similarity := 1.0 - float64(ld(chunk, part, false))/float64(max(len(part), len(chunk)))

			// Update most similar chunk if necessary
			if similarity > maxSimilarity {
				maxSimilarity = similarity
				mostSimilarChunkStart = i
				mostSimilarChunkEnd = i + length
			}
		}
	}

	if maxSimilarity < similarityThreshold {
		return ""
	}

	modifiedWhole := append(wholeLines[:mostSimilarChunkStart], append(replaceLines, wholeLines[mostSimilarChunkEnd:]...)...)
	modifiedWholeStr := strings.Join(modifiedWhole, "")

	return modifiedWholeStr
}

func findSimilarLines(rawSearch, rawContent string) string {
	bestRatio := 0.0
	bestMatch := []string{}
	threshold := 0.6
	bestMatchLineIdx := 0
	search, searchLines := split(rawSearch)
	_, contentLines := split(rawContent)

	for i := 0; i < len(contentLines)-len(searchLines)+1; i++ {
		chunkLine := contentLines[i : i+len(searchLines)]
		chunk := strings.Join(contentLines[i:i+len(searchLines)], "")

		ratio := 1.0 - float64(ld(strings.Join(searchLines, ""), chunk, false))/float64(max(len(search), len(chunk)))
		if ratio > bestRatio {
			bestRatio = ratio
			bestMatch = chunkLine
			bestMatchLineIdx = i
		}
	}

	if bestRatio < threshold {
		return ""
	}

	if bestMatch[0] == searchLines[0] && bestMatch[len(bestMatch)-1] == searchLines[len(searchLines)-1] {
		return strings.Join(bestMatch, "")
	}

	N := 5
	bestMatchEnd := min(len(contentLines), bestMatchLineIdx+len(searchLines)+N)
	bestMatchLineIdx = max(0, bestMatchLineIdx-N)

	return strings.Join(contentLines[bestMatchLineIdx:bestMatchEnd], "\n")
}

// ld compares two strings and returns the levenshtein distance between them.
// refer from https://github.com/spf13/cobra/blob/main/cobra.go#L165
func ld(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				m := d[i-1][j]
				if d[i][j-1] < m {
					m = d[i][j-1]
				}
				if d[i-1][j-1] < m {
					m = d[i-1][j-1]
				}
				d[i][j] = m + 1
			}
		}

	}
	return d[len(s)][len(t)]
}
