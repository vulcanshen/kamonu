# kamonu 開發者備忘

kamonu 是畫 terminu 家族 mark（家紋）的 TUI app，遵循 [terminu design principle](https://github.com/vulcanshen/terminu)（tdp v0.9.1），
建在 terminu-go 上。

## 運作方式

- **目錄**照 tdp 的 project：`cmd/kamonu/main.go` 只管參數（`version`、`help`）；app 本身在 `internal/app`；版本在
  `internal/version`（`kamonu version` 與 splash 印 `version.Display()`）。下面的檔名都在 `internal/app` 裡。
- **畫面交給 lib**（`screen.go`）：這些都是 terminu-go 畫的。
  - `[1] KAMON`（`terminu.List`）與 `[2] PREVIEW`（kamonu 自己的 `Body`）。
  - `NEW KAMON`、`EDIT: <名字>`（lib 的表單）。
  - delete 的 confirm、匯出的 toast、splash、Space menu、key reference、footer。
- **畫格子的 popup 放在 lib 的 popup 疊裡**（`editor.go`，terminu-go v0.1.9 的 `App.Popup`）：
  - kamonu 的 `editor` 是 `PopupBody`：畫框裡的東西（格子、色標、動作列）、處理按鍵、依狀態給 hint 與 key reference。
  - 框、標題、開關動畫、淡化、層色歸 lib。kamonu 沒收下的鍵也歸 lib：沒改過時的 `Esc` 關掉它、`?` 開 key reference、
    `Ctrl-C` 進離開流程（`q` kamonu 收下，不作用）。
  - 上面再開的 confirm（discard、clear、離開）、`SAVE AS` 文字輸入、error popup、key reference 都是 lib 的。
  - 沒存時離開的 confirm 多的那句 Peach，由 `editor.QuitNote` 給 lib。
- **`app.Run` 呼叫 `App.Run`**（`run.go`）：`terminu.Probe` 與輸出裡 `\x1b[K` 的處理都是 lib 的。
  - v0.1.1 以前 kamonu 自己跑 Bubble Tea 的 program、自己畫一疊 popup：那時 lib 的 popup 疊不到 app 自己畫的 popup 上面（紀錄 KM2）。
- **mark**（`mark.go`）：
  - `kamon`：yaml 裡存的東西，包括要不要 U 框、有哪些字、每一格是第幾個字。
  - `mark`：splash 畫出來的樣子，只有底、U 框、金色三種格子。
  - U 框從 `u-frame.svg` 讀進來，編進執行檔。
  - 讀 yaml 時，格子不合規定一律算錯（見「存檔格式」）。
- **存檔**（`store.go`）：
  - `Save`、`Edit` 與匯出都先寫到同一個目錄的暫存檔再 rename，寫到一半失敗時原檔還在。
  - 新增（`Create`、`Save As`）用 `O_EXCL`，不會蓋掉同名的檔。
  - 錯誤訊息不帶路徑（`why`）：使用者知道是哪一個 mark，而路徑可能只是暫存檔的名字。
- **測試守著的**（`kamonu_test.go`、`tdp_test.go`）：
  - `tdp_test.go` 用 lib 的 tdptest（terminu-go v0.4.0）：`TestTDP` 照第一次啟動、`TestTDPWithMarks` 放了兩個 mark，在 tdptest
    的每一種尺寸與 icon 寬度走過每個 panel 的 Space menu、global operation popup、App Log、key reference，再縮到最小尺寸以下又拉回來；
    每個 frame 驗 tdp L4 與有沒有控制字元送到終端機。`TestMoveAside`：啟動時讀不了的 mark 一個一個問，`Enter` 放到 `.bak`，`Esc` 離開。
  - `harness_test.go` 是不用終端機就能跑整個 app 的工具：送按鍵、跑完動畫、讀畫面。超過 200ms 的計時器（toast 留在畫面上的時間）不等，所以 toast 會留在畫面上讓測試讀。
  - tdp L4 與 E5：80×40、120×50、60×30、56×28，加上最小尺寸 56×16（畫格子的 popup 太寬又太高）、120×16（太高）、56×50（太寬：
    六個字的 mark，icon 兩格時色標比格子寬）；icon 與 powerline 符號都一格、icon 兩格、兩種都兩格。主畫面和 kamonu 開的每一種 popup
    （error popup、`DRAW`、它的 key reference、discard 與 clear 的 confirm、`SAVE AS`、離開的 confirm、Space menu、global operation
    popup、`NEW KAMON` 與它的 `Name`、`EDIT: <名字>`、toast、匯出的 menu 與 `EXPORT AS`、delete），每一列都剛好等於終端機寬。
    三種寬度依序跑，同一種寬度的七種大小並排跑（寬度是 lib 全程式共用的一個值）。單獨量每一種 popup 的框歸 lib（tdp v0.8.0 的 E8）。
  - 最小尺寸以下（55×40、80×15、80×5、20×3、1×1）：不當掉、每一列滿寬、畫出名稱與需要的尺寸，按鍵不作用，拉大就回到原本的畫面。
  - `[2] PREVIEW`：`Enter` 放那個家紋、`e` 畫它，沒有東西時都不出現；比 mark 矮時捲動、寫 `N-M of 21`，放得下時不寫、也不列捲動的鍵。
  - 畫格子的 popup 放不下時，下框寫 cursor 那一列 `N of 19`，在按鈕上不寫。
  - 畫格子的 popup 自己畫的那一塊：不管 lib 給多窄、多矮，每一列都剛好是給的寬，列數不超過給的高。
  - U 框跟 tdp 的 `principle/icon.svg` 一模一樣（`internal/app/testdata/tdp-icon.svg` 是它的副本，只給測試用）。
  - tdp E6：splash 畫的 `internal/app/icon.svg` 跟 `docs/icon.svg` 一模一樣，lib 的 splash 吃得下；它的 U 框是家族的，註解依序是 U、K、A、M、O、N。
  - yaml：寫出去再讀回來不變；手改出錯一律回錯誤，不會讀成另一張圖。
  - 匯出的 svg：註解的順序是 U 框再照字的順序；沒有格子的字不寫；字全是金色；lib 的 splash 吃得下。
  - 改字、打開 U 框之後格子的變化（見「改資料」）。
  - 畫格子：cursor 跳過 U 框、到邊停住；`Space` 照順序切換；色標每個字一個顏色；改過才問，`Esc` 回到格子。
  - Save 與 Save As：名字的檢查、控制字元、原檔不動、cursor 停在新檔。
  - 存檔失敗開 error popup，`Enter` 關掉後回到按鈕；打開的動畫播完前不收鍵；`Ctrl-C` 兩次離開。
  - 匯出寫出跟 yaml 一致的 svg，失敗時用 toast 說原因；壞掉的 yaml 在預覽、Draw、Edit 都說讀不了。
  - 已經有 svg 時開 menu：`Overwrite <名字>.svg` 蓋掉原檔；`Export as new…` 開輸入框，說明列舉出下一個號碼，自己取的名字照用，
    已經有的擋下（加不加 `.svg` 都一樣）；`Esc` 什麼都不寫；寫檔失敗開 error popup。

## 設計決定

### 畫面

- **單一畫面、兩個 panel**：
  - `[1] KAMON` 是資料目錄裡 `conf/*.yaml` 的清單，名字不顯示 `.yaml`。
  - `[2] PREVIEW` 畫 cursor 那一筆，上下各多一列底，跟 splash 一樣是 21 列 × 50 欄。
  - 理由：kamonu 只做一件事，挑一個 mark、打開來畫。
- **最小尺寸 56 × 16**（2026-10-20，tdp v0.8.0 的 L1；lib 的 `App.MinWidth`、`MinHeight`）：比這個小，lib 只在中間畫名稱與
  `needs a terminal of at least 56×16`，按鍵都不作用（連按兩次 `Ctrl-C` 離開），拉大就回來。
  - 寬 56：畫格子的 popup 要 54 欄（格子 50 欄、左右各一格留白、框），popup 左右各留一欄（tdp F7）。再窄格子就被截掉，cursor
    可能跑到看不見的地方。
  - 高 16：popup 最高是畫面減兩列，扣掉框、上下留白、色標、分隔線、按鈕，格子還看得到六列，跟著 cursor 捲。
  - 其他畫面都比它小：表單、menu、key reference 放不下時照常捲；預覽放不下時捲；splash 在 50 × 30 以下只畫字。
- **名字太長時從尾端截、加 `…`**（2026-10-20，tdp v0.8.2 的 layout/panel）：`[1]` 只有一欄 Name，標成長欄（`Column.Shrink`），比 `[1]` 寬的名字照欄的截法截。
- **`[1]` 的 label 是 `KAMON`**（不是 `ICONS`）：user 定的，清單裡的東西就是家紋。
- **`[2]` 至少 54 欄**（mark 50 欄、左右各一格留白、框兩格）。80 欄時 `[1]` 剩 26 欄；更窄時一次只畫一個 panel。
- **預覽照 splash 的樣子，字全是金色**（2026-10-14，user）。理由：kamonu 裡只有這裡看得到成品的樣子；想看分字，按 `Enter` 打開畫格子的 popup。
- **`[2] PREVIEW` 只有一個項目，就是它畫的那一個 mark**（user 2026-10-19，紀錄 KM8、KM17）：
  - `Enter` 是 Show（名字 user 2026-10-20 確認）：把那個家紋照它在 app 裡呈現的樣子放出來（家族 app 按 `V` 的那個效果），用 lib
    的 `App.PlaySplash`（terminu-go v0.2.0）播 `terminu.NewSplash(km.svg(), <名字>, "", "")`，hint `Enter:show`，任何鍵關掉。
    讀不了的 kamon 開 error popup `SHOW FAILED: <名字>`。v0.2.0 以前 lib 放不了 app 指定的 mark，`Enter` 暫時是 Draw。
  - **mark 下面只寫名字**（user 2026-10-20 確認）：app 的 splash 下面寫名稱、版本、介紹；kamonu 是家徽產生器，不管 app 本身，
    所以版本與介紹空著，作者、信箱、提示照 lib 畫。
  - item operation 是 `[e]dit`：打開畫格子的 popup。`[1]` 的 Edit（改名字、字、U 框）也是 `[e]dit`，兩個同名，靠所在的 panel 分
    （user 2026-10-19，紀錄 KM17）。中間改定過 `[d]raw`，撤回：`d` 是 tdp K12 保留的半頁鍵，`[2]` 放不下時也用它捲。
  - 沒有東西可以預覽時沒有項目，兩個鍵都不出現，下框也沒有 hint（tdp v0.8.0 的 layout/panel：沒有按得動的鍵就不寫 hint，紀錄 KM14）。
  - 2026-10-18 重構時暫定過 `Enter` 是 `[2]` 的 panel operation、只做 Draw。
- **預覽放不下時捲動**（2026-10-18，components/layout/panel 的「沒有 cursor 的 panel」）：mark 21 列，`[2]` 比它矮時（終端機不到
  24 列）用 lib 的 `Pager` 捲，下框右邊寫 `N-M of 21`，`j|k`、`u|d`、`gg|G` 只在這時候有、也只在這時候列進 key reference。
  寬度不夠（窄寬時 `[2]` 不到 52 欄）從右邊截掉。
- **預覽從左上角畫起**（2026-10-18，components/layout/panel 的框：內容離左框一格、第一列緊接上框）：以前上下左右置中，像 splash；
  照 panel 的規定改成跟其他 panel 一樣從左上角開始。
- **預覽的空狀態與讀不了**（2026-10-18，layout/screen 的空狀態、layout/panel 的錯誤）：
  - 清單是空的：`Nothing to preview`，提示 `Press [1], then [A] to add one`（`[A]` 是 `[1]` 的鍵，先到 `[1]`）。
  - yaml 讀不了：事實 `Can't draw <名字>` 用 Red（`[2]` 的 `Failure`），提示 `Press [Enter] to see why` —— `Enter` 打開的
    error popup 寫著原因。以前兩句都是 Overlay0，第二句是原因本身。
- **`[1]` 的空狀態提示是 `Press [A] to add one`**（2026-10-18）：原本還有 `, or [Space] to see what you can do here`，80 欄時
  `[1]` 內寬 24，被 lib 截到貼著框（紀錄 KM11）；縮短後放得下。`Space` 本來就寫在 footer。tdp v0.8.0 定了空狀態放不下時折行
  （S12），lib 做了之後不必再為了寬度縮短；現在的句子也夠用，照留。

### 動作

清單上的動作（2026-10-14，user）：

| 鍵 | 動作 | 範圍 |
|---|---|---|
| `Enter` | Draw：打開畫格子的 popup | item |
| `e` | Edit：`EDIT: <名字>` 表單，改名字、字、U 框（`[e]dit`，user 2026-10-19，紀錄 KM17；以前是 `E`） | item |
| `P` | Export：寫出 svg | item |
| `X` | Delete：先 confirm（lib 的 `Action.Ask`：從 Space menu 開時 menu 留在底下，`Esc` 回到 menu，tdp T1、F4） | item |
| `A` | Add：`NEW KAMON` 表單 | panel |

- **`Enter` 直接畫格子，`[e]dit` 改資料**：
  - 畫格子是最常做的事，按一下就到。
  - 忘了鍵的話，`Space` menu 裡全列著。
  - tdp K3 的 `Enter` 是對這一筆最直觀的動作；先選再做是 Space menu 的事。
  - 規範 menu 範例（components/dialog/menu）也是這個排法：`[Enter] Connect`、`[e]dit`。
- **`Enter` 原本叫 Edit**：有了改資料的 Edit 之後改名叫 Draw，免得兩個動作都叫 Edit。
- **原本的 `[r]ename` 併進 Edit**：改名就是改 Name 那一欄。
- **匯出用大寫 `P`**：
  - 不用 `x`：規範 menu 的範例裡刪除是 `[X] Delete`，用過家族 app 的人會把 `x`／`X` 當成刪除。
  - user 選大寫 `P`，因為瀏覽器的列印是 `Cmd-P`。
  - label 是 `Export`。tdp v0.8.0 的 M5：熱鍵的字母出現在 label 裡就原地包，不分大小寫，括號裡照實際的鍵 —— `Ex[P]ort`
    （紀錄 KM16，terminu-go v0.2.0）。以前的 lib 比對時大小寫算數，為了讓它原地包，label 寫成 `ExPort`，key reference 裡就看到
    `ExPort` 這個字。
- **`[X] Delete`**：
  - 先 confirm（`Delete <name>?`），上面一句 Peach 說 `conf/` 裡的 yaml 會被刪掉。
  - 用大寫 `X`（2026-10-14，user 選「跟規範」）：規範 menu 的範例（components/dialog/menu）就是 `[X] Delete`。原本是 `[D]elete`。
  - 匯出過的 svg 不跟著刪：那是已經給出去的成品。

### 新增：`NEW KAMON` 表單

- **只有家族的大小 25 × 19**（2026-10-14，user）：kamonu 是 terminu 家族家徽的編輯器，不是一般的點陣圖編輯器。
- 欄位：
  - `Name`
  - `Letters`：框裡有哪些字（見「分字」）
  - `Base`：一個 checkbox `U frame`，預設勾著，決定要不要 U 框
- **按 `Create` 當場建檔**，不是畫完存檔才建（2026-10-14，user）。理由：名字一開始就定，檔案一開始就在，之後只是改它。建好之後打開畫格子的 popup。
  `NEW KAMON` 收起與 `DRAW` 打開同時播，`DRAW` 從第一格起就是第 1 層色，不跳色（tdp v0.6.10 的 T1 與 components/color 的層色）。

### 改資料：`EDIT: <名字>` 表單

- lib 的表單，三欄跟 `NEW KAMON` 一樣，帶入現在的值（灰字提議是舊值），按鈕是 `Save`。
- **可以改，後果自付**（2026-10-14，user）：
  - **字照位置對應**：第 1 個字的格子永遠是第 1 個字。
    - `FIL` 改成 `FAL`：中間那組格子形狀不動，改標成 A。
    - `KB` 改成 `FIL`：K 的格子變成 F、B 的變成 I；L 沒有格子，要自己進去畫。
  - **字數變少**時，沒了的那幾個字的格子變回底。
  - **U 框原本沒有、打開它**時，畫在框位置上的格子清掉。
  - 都不先問（後果自付是 user 定的原則；後兩條是照這個原則補的預設）。
- **標題是 `EDIT: <名字>`**，例如 `EDIT: filu`：照 tdp v0.6.1，名字接在 `: ` 後面、大小寫照原樣（紀錄 KM5，lib 的 `Form.Subject`，
  terminu-go v0.1.1）。discard 的問句也帶名字（`Discard changes to filu?`）。
- 改名時，名字跟原本一樣不算重複；換成別的名字才檢查有沒有被用掉。

### 分字

- **建立時先填框裡有哪些字**（2026-10-14，user）：例如 filu 填 `FIL`（U 是框）。
  - 字的順序就是 splash 揭露的順序。
  - 同一個字可以出現兩次：sshu 是 S、S、H，兩個 S 各算一個字。
  - 只收 A 到 Z，小寫自動轉大寫。
- **最多 6 個字**（2026-10-14，user 選 A）：
  - U 框裡能畫字的地方是 15 欄 × 17 列（第 5–19 欄、第 0–16 列）。
  - 認得出是字母的最小大小是 3 × 5，字與字之間空一格。字貼著框塞滿是 4 × 3 = 12 個；跟框留空是 3 × 2 = 6 個。
  - catppuccin 的 14 個彩色去掉藍色系（會跟 U 框、cursor 混在一起）剩 10 個，一眼分得開的大約 7 個。
  - 家族最長的是 kamonu 的 5 個字。
- **每個字一個顏色，照順序固定**：Peach、Teal、Mauve、Green、Red、Yellow。
  - 顏色只在 kamonu 裡用來分字；匯出的 svg 照 tdp S5，字一律是金色。
  - 冷暖交錯排，順序相鄰的字通常也畫在隔壁，這樣顏色差最多。
  - user 說顏色由 kamonu 決定、分得出來就好；不每次隨機，同一個 mark 每次看起來一樣才認得。
  - 規範：內容本身的顏色由 app 決定（components/color 開頭、tdp P4 最後一段）。格子裡的字是內容。

### 存檔格式

- **yaml，存成一張圖**（2026-10-14，user 選 A）：
  ```yaml
  frame: true
  letters: [F, I, L]
  cells: |
    .........................
    .......111.222.3.........
    ...（共 19 列）
  ```
  - `cells` 一列一行，`.` 是底，`1` 到 `6` 是第幾個字。用數字不用字母，是因為同一個字可能出現兩次。
  - U 框不畫進 `cells`，它是 `frame` 這個設定，匯出時才補上（user）。
  - 理由：打開檔案就看得到畫的樣子，改一格的 git diff 也只差一行。
- **手改出錯一律回錯誤**，不會讀成另一張圖：
  - 縮排錯了，yaml 自己擋（tab：`found a tab character…`；少一格：`did not find expected key`）。
  - yaml 讀得進來的錯（縮排多一格、少打一格、行尾多空白），那一列都不是 25 格，kamonu 擋：`Row 8 has 24 cells, not 25`。
  - 其他：超過字數的數字、畫在 U 框上的格子、字不是一個 A 到 Z、列數不是 19、不認得的欄位、空檔案。
  - 列與欄從 1 數，寫給人看。
- 否決的 B（每個字列出自己的格子座標）：打錯一個數字（`[8,6]` 打成 `[8,7]`）還是合法的 yaml 與座標，那一格會默默跑到別的地方。

### 匯出

- **`Ex[P]ort` 直接寫到 `svg/<名字>.svg`**，再用 toast 說寫到哪裡（2026-10-14，user 選 B）。toast 是 Success（綠色，tdp v0.8.0
  的 components/dialog/toast，紀錄 KM15），寫完整路徑，長了就折行（v0.2.0 的 toast 不再限三列）。
  - 失敗時開 error popup `EXPORT FAILED: <名字>`（2026-10-18）：原本用 Error toast，可是那時 toast 最多三列，原因長時會被截掉，
    完整的內容要走 error popup 或 log，kamonu 沒有 log。
- **已經有 `<名字>.svg` 時先問：menu 加輸入框**（user 2026-10-19，紀錄 KM10）：
  - menu `EXPORT` 兩列：`Overwrite <名字>.svg` 直接蓋掉；`Export as new…` 開輸入框 `EXPORT AS`，輸入新檔的名字（值叫 `File`，
    打不打 `.svg` 都可以），已經有的名字在輸入框裡就擋下（`File is taken: …`）。
  - 為什麼不是表單：只有一個值（新檔名），照 components/dialog/form 該用 input popup；蓋掉用不到這個值，是另一種做法，所以先選做法。
  - 2026-10-15 到 2026-10-18 是表單 `EXPORT: <名字>`：一欄 `File` 加 `Overwrite`、`Export` 兩顆按鈕。
  - menu 的標題是 `EXPORT: <名字>`，名字照原樣（lib 的 `MenuAbout`）。
  - 輸入框用建議值帶入 `<名字>_<n>.svg`（n 是從 1 起第一個還沒用掉的號碼；tdp v0.8.0 的 components/input），不碰按 `Enter`
    就用它。
  - 輸入框上按 `Esc` 回到 menu（tdp v0.8.0 的 T1：取消一律回到 menu）；寫好了整疊關掉，跳 Success toast。寫檔失敗時輸入框留著、
    打的名字還在，上面開 error popup（lib 的 `TextInput.Commit`）。
- **格式照 tdp S5**：
  - 底 `#000000`。
  - `frame: true` 時寫 `<!-- U -->` 加上 U 框。
  - 接著照字的順序，每個字一個註解（`<!-- F -->`），後面是它的格子，全部金色 `#f2b753`。
  - 格子照列的順序寫，tdp S5 的揭露動畫會把同一個字的格子打散。
- **沒有格子的字不寫**：splash 沒有東西可以揭露，空的一段只會多停一下。
- 改名、刪除都不動 `svg/` 裡匯出過的檔。

### U 框

- **U 框是一個獨立的檔案**，從家族 mark（tdp 的 `principle/icon.svg`）的 U 那一部分抽出來，放在 repo 的 `internal/app/u-frame.svg`、編進執行檔：使用者碰不到、刪不掉。tdp S5 規定 U 框全家族一樣、不改。
- **mark 只記要不要 U 框**（yaml 的 `frame`，2026-10-14，user）：U 框的格子不存在 mark 裡，畫不到、刪不掉。

### kamonu 自己的 mark

- **`docs/icon.svg` 是 user 用 kamonu 畫、匯出的**（2026-10-14）：U 框裡 K、A、M、O、N。splash 畫的是它（tdp E6），換掉了原本借用的 tdp mark。
- **程式裡另有一份 `internal/app/icon.svg`**（tdp v0.7.0 的 project）：`go:embed` 讀不到 `docs/`，splash 編進執行檔的是這一份，
  `TestIconIsTheSplash` 守住兩份一模一樣。`docs/icon.svg` 是正本：改它時把它複製到 `internal/app/icon.svg`。
- **repo 只放 `docs/icon.svg`，不放 yaml**（2026-10-14，user）：`docs/icon.svg` 是家族的慣例，yaml 是 kamonu 自己的東西。原稿在
  user 的 `~/.kamonu/data/conf/kamonu.yaml`。

### 畫格子的 popup

- **kamonu 自己的 popup**（2026-10-14，user：「這個屬於 kamon 自己的 popup，lib 不會有這個東西，也不應該有」）：
  - 不是表單加一種欄位，在 tdp F1 七類以外（見「偏離 tdp」）。
  - 框裡的東西與按鍵由 kamonu 做；框、標題、hint 的位置、開關動畫、淡化、層色由 lib 畫（terminu-go v0.1.9 的 `App.Popup`）。
  - 上面再開的 discard confirm、`SAVE AS` 文字輸入、error popup 都用 lib 的。v0.1.1 以前這些都由 kamonu 自己畫（見「運作方式」）。
- **按 `Enter` 或 `Space`，那一格依序切換**：底 → 第 1 個字 → … → 最後一個字 → 底（tdp K5 的 checkbox 例外）。
  - 格子下面一列色標，每個字寫成 `[1] ■ F`：按哪個數字、什麼顏色、哪個字。數字照 tdp M5 標成熱鍵（數字放前面）。
    方塊與字之間一格、字與字之間兩格（components/README 的間距）。popup 照色標放得下的寬度開（6 個字、icon 佔兩格要 58 欄）；
    終端機太窄、放不下一列時，放不下的字接到下一列（2026-10-18；以前把間距縮成一格、方塊與字黏在一起）。
  - 下面是動作列，`Tab` 在格子與動作列之間切換，跟表單一樣。
- **放不下時格子捲動**（2026-10-18）：整個 popup 要 27 列（格子 19 列、空一列、色標、分隔線、按鈕，加上下留白與框）。
  - 終端機 28 列時少一列：拿掉色標上面那一列空白。
  - 更矮時（例：80 × 24）照表單的做法：只有格子捲，捲最少的量讓 cursor 那一列看得到；色標、分隔線、按鈕固定在下面
    （components/dialog/form、layout/popup 的「內容放不下時」）。在按鈕上時格子停在原處。
  - 下框右邊寫捲動位置：cursor 在第幾列，`<列> of 19`（lib 的 `PopupBody.Scroll`，紀錄 KM9）；在按鈕上不寫，跟表單一樣。
  - 會捲時 popup 照「hint 加上最寬的捲動位置 `19 of 19` 再加 9」開（layout/popup 的寬度：捲的時候框不忽寬忽窄），內寬 55。
- **數字鍵 `1`–`6` 直接把那一格設成第幾個字**（2026-10-14，user：「連續按 space 太繁瑣」）。
  - **再按一次同一個數字就清空**（2026-10-14，user）：已經是那個字的格子上按它的數字，等於 `Backspace`，畫錯時不用換鍵。
  - 超過字數的數字不作用（`FIL` 按 `4`）。
  - popup 開著時按鍵只給 popup（tdp K 的「popup 能按的鍵」），所以不會碰到畫面上 `1`–`9` 切 panel。
- **`Backspace` 把那一格清成空的**，不管它是第幾個字（2026-10-14，user）：不用一路 `Space` 繞回底。
- **`p` 切換成金色**（2026-10-14，user）：格子與色標的字全畫成金色，就是匯出的 svg 與 splash 的樣子；再按一次變回分字的顏色。
  - 為什麼：user 用分字的顏色畫好 kamonu 自己的 mark，匯出才發現字黏在一起 —— 顏色不同時分得開，全是金色就成了一整塊。
    右邊的預覽是金色的，但畫的時候被 popup 蓋住。
  - 金色時照樣可以畫；每次打開都從分字的顏色開始。
  - 用 `p`（preview）：tdp 裡 `v` 是選取模式（tdp K11）、`V` 是 splash（tdp S1），`p` 沒有用到。
- **`x` 清空所有格子，先 confirm**（2026-10-14，user）：`Clear every cell of <名字>?`，`Enter` 清空、回到格子，`Esc` 不動。
  - 一格都沒畫時不問，什麼都不做。
  - 清空只改 popup 裡的圖，`Save` 才寫進檔案；改過了，`Esc` 離開時照樣先問要不要丟掉。
  - 用小寫 `x`：清單上刪除 mark 是大寫 `X`（規範 menu 的範例），意思相近（刪掉東西），大小寫分開不會按錯地方。
- 數字鍵、`Backspace`、`p` 與 `x` 只寫在 `?` 的 key reference 裡（數字另外寫在色標上）：hint 放不下，popup 只有 54 欄，再加會把
  `Esc:cancel` 擠掉。數字鍵與 `Backspace` 在按鈕上不作用；`p` 與 `x` 在哪裡都可以按。
- **按鈕是 `Save`（存回原檔）與 `Save As`（存成新檔）**。存完就關掉，跟表單送出一樣（2026-10-14，user）。
  - `Save`：寫回原檔，回到清單。
  - `Save As`：先開 `SAVE AS` 文字輸入填新名字（檢查跟 `NEW KAMON` 一樣），連同字與 U 框存成新檔，原檔不動，關掉整疊，清單的 cursor 停在新檔。
  - 理由：按鈕就是送出，家族的表單送出就關。
- **U 框的格子是停用的格子，cursor 跳過**（2026-10-14，user）：
  - 照日曆範圍外的日子：停在那個方向下一個能停的格子，那個方向沒有了就不動。
  - 理由：家族裡停用的東西 cursor 一律跳過（menu 的停用列、表單的停用欄位、日曆）。
  - 沒有 U 框的 mark，每一格都能畫。
- 格子裡 hjkl 上下左右，到邊停住（tdp K12：照空間走的不繞）。沒有錯誤列：格子不會不合格，存檔失敗開 error popup。改過時 `Esc` 先問。
- cursor：層色底，方塊保留自己的顏色，格子的顏色就是它的值（tdp v0.6.7 components/color 的 cursor 色塊例外，紀錄 KM3）。
- 下面這些是寫程式時照預設定的（2026-10-14，跟 user 報告過）：
  - 標題 `DRAW: <名字>`，glyph `nf-md-grid`（U+F02C1）。打開時 cursor 在第一個能停的格子。
  - 按鈕 `Save As`、`Save` 並排靠右。`Tab` 第一次到動作列時停在 `Save`，之後回到離開前的那一個。
  - 按鈕上 `h|j|k|l` 移到前一個、後一個，到頭尾繞回；`gg|G` 到第一個、最後一個（tdp v0.6.7 dialog/form 的動作區）。
    v0.1.0 只有 `h|l`、不繞：那時規範的動作列只有一個按鈕，兩個按鈕之間怎麼移沒寫（紀錄 KM1）。
  - hint：在格子上是 `Enter|Space:cycle Tab:actions Esc:cancel`，在按鈕上是 `Enter:save Tab:grid Esc:cancel`（或 `Enter:save as`）。
    按鈕上寫 `Tab:grid`，不是表單的 `Tab:fields`（user 2026-10-14，經 terminu 的 session）：`Tab` 的 hint 寫要去的那一塊
    （components/dialog/form），這裡要去的是格子；`Tab:fields` 是表單與有按鈕的 input popup 的字，那裡要去的是欄位。
    tdp v0.6.5 的 K2 只管 `Tab` 在欄位區與動作列之間來回，格子 popup 本來就是這樣。
  - `q` 在這個 popup 上不作用（照表單：`q` 不進離開流程；這是偏離，見「偏離 tdp」）。`Ctrl-C` 開離開的 confirm；畫的東西沒存時，多一句 Peach `Changes to <名字> are not saved.`。
  - `SAVE AS` 沒有灰字提議（沒有舊值）；說明列是 `Name for the new kamon; <名字> stays as it was`。什麼都沒打就按 `Enter`
    會擋下來寫 `Name can't be empty`（tdp v0.8.0 的 K3：打開時是空的、又不能空的值）；v0.1.2 到 v0.2.0 以前是直接關掉。寫檔
    失敗時輸入框留著、上面開 error popup（`TextInput.Commit`）。名字有控制字元時，錯誤列寫 `Name can't have control characters`：錯誤點名的是值的名字，
    標題 `SAVE AS` 說的是在做什麼，所以 kamonu 另外給值的名字（tdp v0.6.11 的 components/input、terminu-go v0.1.10 的
    `TextInput.Name`；紀錄 KM7）。

### 名字

- 檢查：
  - 不能空：`Name can't be empty`。
  - 不能有 `/`、不能 `.` 開頭：這兩條是名字本身的格式，在輸入框裡就擋。
  - 不能跟現有的重複：`Name is taken`，表單在送出時擋。
  - 不能有控制字元：lib 的輸入在 `Enter` 時擋。
  - 前後的空白字元會去掉。
- 清單照名字排，不分大小寫。

### 離開

- 主畫面按 `q` 先問（`Quit kamonu?`）。理由：家族的 app 都問；畫到一半的東西在畫格子的 popup 裡，那裡另外會提醒沒存。

### 位置

- 資料目錄 `~/.kamonu/data/`（tdp E3），`KAMONU__DATA` 覆寫（tdp E2）。裡面分兩個子目錄（2026-10-14，user）：
  - `conf/<名字>.yaml`：mark 本身。
  - `svg/<名字>.svg`：匯出的成品。
- yaml 算資料不算設定：tdp E3 的設定目錄（`~/.config/kamonu`）放的是決定 app 怎麼運作的設定檔，yaml 是使用者畫出來的東西。
  tdp v0.9.0 的「刪掉會怎樣」：mark 與匯出的 svg 刪了就找不回來，都是資料；kamonu 沒有設定、狀態、快取，位置也沒換過，不用搬。
- **讀不了時**（2026-10-20，tdp v0.9.0 的 E3；lib 的 `App.MoveAside`，mark 用 v0.4.2 的 `App.MoveAsideRecord`）：啟動時（`App.Start`）先讀 `conf/` 與每一個 mark，第一個讀不了的
  開 `MOVE ASIDE`（路徑與原因，`Enter:move Esc:quit`）。`Enter` 把它改名成 `<名字>.yaml.bak`（已經有就 `.bak.1`……），記進 App Log，
  再檢查下一個；`Esc` 離開，路徑與原因印到 stderr。一次問一個，每一個都說得出是哪個檔。
  - 改名後不另外建一個新的：mark 是一張一張的檔，放到旁邊就從清單上消失；空的 mark 沒有字，建不出來。tdp v0.9.1 照這樣定了
    （E3：資料一筆一個檔時，問句寫 `Move it to <檔名>.bak and go on without it?`，紀錄 KM18）。`conf/` 目錄本身讀不了時照舊用
    `MoveAside`（問句寫 `and start a new one?`）：放到旁邊後清單是空的，新增時再建。
  - 開著時才壞掉的 mark（例：在外面手改了）照舊：預覽寫 `Can't draw`，Draw、Edit 開 error popup。
- **App Log**（2026-10-20，tdp v0.9.0；lib 給的，在 panel 上按 `!`，footer 常駐 `!:log`）：toast、error popup、放到旁邊的檔由 lib 記；
  kamonu 自己記做完的事：`created <名字>`、`saved <名字>`、`saved <名字> as <新名字>`、`renamed <舊> to <新>`、
  `saved <名字>: letters …, U frame on|off`、`deleted <名字>`。

## 已否決，不要重提

- 同一個動作同時綁 `Enter` 與 `e`：menu 裡會出現兩列一樣的動作。
- **full size 與自訂大小**（2026-10-14）：曾想讓新增時選 locku 鎖定畫面的大小或任意大小。否決，kamonu 只畫家族的 mark（25 × 19）。
  附帶查到的：locku 的鎖定畫面沒有固定的格數，是 `終端機寬 ÷ 2` × `終端機高 − 1`。
- 存檔時才輸入檔名、才建檔：改成新增時就填名字、當場建檔。
- 存完留在畫格子的 popup 裡繼續畫：存完要另跳 toast 才知道存了，按鈕的意思也會從「送出」變成「做一件事、留著」。
- **用 popup 的層色分字**：
  - 只有 4 個，kamonu 自己就有 5 個字。
  - 4 個都是 Lavender 到 Sapphire 之間的淺藍，相鄰的字分不出來。
  - 層色另有用途（popup 的框與 cursor，tdp P4）。
- 字建好就不能改：user 要可以改，後果自付。
- `Enter` 先開一個選單（`Edit kamon`／`Edit properties`）：每次畫格子都多一步，而且跟 Space menu 重複。
- `Enter` 先開資訊 popup，動作列有 `Edit Kamon` 與 `Save`：表單兩個按鈕（KM1）；畫格子的 popup 疊不到 lib 的表單上；兩個 `Save` 意思不清。
- 匯出用 `x`：跟家族的刪除鍵 `X` 只差一個 Shift。
- **已經有 svg 時直接蓋掉**（v0.1.0 的做法）：user 要先問，舊的成品可能還要留著。
- **已經有 svg 時用 menu 問**（`[N]ew file`、`[O]verwrite`，2026-10-15）：user 選表單的多個按鈕。
- **一句話加兩個按鈕**（`filu.svg is already in …`，下面 `Overwrite`、`Save as filu_1`）：confirm 沒有按鈕、表單要有欄位，
  規範沒有這種 popup，要補規範與 lib；user 選有 `File` 欄位的表單，現在就做得出來。
- 匯出時用 lib 的「選目錄」挑地方、寫成那裡的 `icon.svg`：user 選固定寫到資料目錄（按一下就好），要放進 app 的 `docs/` 時自己複製。
- yaml 放在 `~/.config/kamonu`：照 tdp E3，畫出來的東西算資料。
- 預覽用分字的顏色：kamonu 裡就看不到金色的成品了。
- **用那個字母本身當熱鍵**（按 `k` 畫 K，2026-10-14）：小寫撞到格子裡的移動鍵（kamonu 的 K 是 `k` 往上、filu 的 L 是 `l` 往右），
  還有 `p`、`q`、`g`；大寫（`Shift-K`）不撞，但 user 選維持只用數字，改成數字可以 toggle。
- 把 kamonu 自己的 mark 的 yaml 放進 repo（`docs/icon.yaml`）：yaml 是 kamonu 的東西，app 的 repo 只放家族慣例的 `docs/icon.svg`。

## 已知的牆與未做

- **當 locku 的保護程式**：locku 讀 kamonu 畫的圖當一種 saver。這是 locku 的功能，locku 改用 lib 之後再談。

## 偏離 tdp

- **tdp F1（popup 只有七類）**，在畫格子的 popup：它是 kamonu 自己的一類（tdp v0.7.x 的 F1 拿它當七類以外的例子）。
  - 為什麼：內容是 25 × 19 的格子，七類都裝不下（最接近的 form 是「一列一個值」）。
  - 長相與按鍵盡量照 form：`Tab` 在格子與動作列之間、按鈕送出、改過時 `Esc` 先問；格子裡照 tdp K12。
- **tdp K5（popup 上的 `Space` 不作用）**，在畫格子的 popup 的格子上：`Space` 把這一格換成下一個字（跟 `Enter` 一樣）。
  - 為什麼：每一格當成 checkbox（tdp K5 唯一的例外是 checkbox 的選項，`Space` 勾或取消勾；2026-10-13 跟 user 定的）。後來一個 mark
    可以有好幾個字，一格不只亮暗兩種，`Space` 改成照順序換字，最後一個字之後回到空的 —— 已經不只是勾或不勾。
  - 在按鈕上 `Space` 不作用，照 tdp K5。
- **tdp K1、K9（`q` 進離開流程）**，在畫格子的 popup：`q` 不作用，跟表單上一樣；`Ctrl-C` 照樣進離開流程。
  - 為什麼：這個 popup 的長相與按鍵照表單（`Tab`、按鈕、改過先問），tdp K1 的例外只寫了「表單上的 `q`」。同一種樣子的 popup，
    `q` 一個會走、一個不會，使用者得記兩套；畫到一半手滑按到 `q` 也不會跳出離開的 confirm。
- **tdp E6、E7（demo gif）**，在 README：沒有 VHS 錄的 demo gif，`.local/demos/` 也沒有腳本（`make gif` 是 family.mk 給的，跑了會找不到 tape）。
  - 為什麼：kamonu 大部分是 user 自己在用，放上 GitHub 但不推廣（user 2026-10-14：「不需要 vhs 錄 gif」）。
  - 其他照 tdp E6：goreleaser、`install.sh`／`uninstall.sh`、Homebrew tap、tdp L4 與 icon 的測試。

## 設計文件導讀

- 規範的盲點與 lib 要補的地方記在 terminu 的 `.local/family-fix/kamonu/README.md`（編號 KM）。
- 還沒符合 tdp 的地方逐條寫在 `kamonu-terminu-fix.md`（tdp E7），全部修完就刪掉；現在沒有這個檔。terminu-go v0.2.0 時修完刪過
  一次；tdp v0.9.1 的 E3（KM18）等 lib 時又有過，v0.4.2 修完刪掉。

## 建置與開發

- **terminu-go**：`go.mod` 指向 GitHub 上打過 tag 的版本（現在是 v0.4.2）。
  - 2026-10-20 的一輪（tdp v0.9.0）lib 還沒打 tag 時，改動用 `go.work` 做、先留在工作區，等 v0.4.0 打了 tag、`go.mod` 改好、整個
    `make check`（含 `check-lib`）過了才 commit：照這裡的規定，不 commit 只有本機 lib 才編得過的程式。
  - 開發時用 `go.work`（`go work init . ../terminu-go`，terminu-go clone 在 kamonu 旁邊）：build、test 用的是本機的 terminu-go，
    lib 一改馬上看得到。`go.work`、`go.work.sum` 不進版控。
  - commit 與發版前，關掉 go.work 跑一次（`make check-lib`：`GOWORK=off go build ./...`、`GOWORK=off go test ./...`）。過不了，
    表示用到了 lib 還沒發版的東西，要等 terminu-go 打新 tag。
  - 升級：`go get github.com/vulcanshen/terminu-go@vX.Y.Z`，再 `make tidy`（`GOWORK=off go mod tidy`），commit `go.mod` 與 `go.sum`。
    發版時 `go.mod` 只能指向 terminu-go 打過的 tag，不用 replace、不用 pseudo-version。
- **`Makefile`** 只設 `BINARY := kamonu`、`include family.mk`（tdp v0.7.0 的 project）：`family.mk` 是 tdp 的範本，照原樣複製，
  不在這裡改；要改先回報給 tdp。`make` 列出所有指令。
  - `make build`：編出 `./kamonu`（`CGO_ENABLED=0`、`-trimpath`、strip，版本是 `dev`）；`make run`：直接跑。
  - `make test`：約一分鐘（每個 popup 的開關動畫照真的時間跑）；`make test-race`：帶 race detector。
  - `make check`：fmt-check、vet、test -race，有 `go.work` 時再加 check-lib，commit 前跑，約兩分鐘。
  - `make install`／`make uninstall`、`make package`（`dist/` 的 tar.gz）、`make clean`。
  - `make release-check`、`make snapshot`：見「發布」。
  - kamonu 自己的 target：沒有（`##@ kamonu 自己的` 是空的）。
- **project 圖上沒有的頂層檔案**：`CLAUDE.md`（跟 AI 協作的交代）。
- **從原始碼裝**：`go install github.com/vulcanshen/kamonu/cmd/kamonu@latest`（v0.1.2 以前 main 在根目錄，是
  `github.com/vulcanshen/kamonu@latest`）；或 clone 下來 `make build`。
- 不想動到自己的 `~/.kamonu`，就用 `KAMONU__DATA=<別的目錄> ./kamonu`。

## 發布

照 tdp 的 project「發布」（tdp E6）：goreleaser 打包 linux／darwin × amd64／arm64，`install.sh`／`uninstall.sh`，Homebrew tap 的
formula。`.github/workflows/release.yml`、`.goreleaser.yaml`、`install.sh`、`uninstall.sh` 都是 tdp 的範本，只改範本表上寫的地方
（formula 的 description 與 caveats、安裝完最後那段提示）。

- **步驟**：
  1. `make check`。
  2. CHANGELOG 的 `## Unreleased` 改成 `## vX.Y.Z — tdp vA.B.C`。
  3. commit、打 tag `vX.Y.Z`、push commit 與 tag。
  4. `.github/workflows/release.yml` 接手：在 ubuntu 與 macos 上跑 `go test -race`；goreleaser 發 release、把 formula 推到
     `vulcanshen/homebrew-tap`；最後用 CHANGELOG 那一段蓋掉 release 的說明（推 tap 失敗時照樣蓋，release 還沒建好就跳過）。
- **release 說明**：workflow 用 awk 取 CHANGELOG 裡 `## <tag> ` 開頭的那一段，到下一個 `## ` 為止。
- **版號**：goreleaser 把去掉 v 的版號注入 `internal/version.Version`（`-X github.com/vulcanshen/kamonu/internal/version.Version=…`），
  `kamonu version` 與 splash 印 `version.Display()`，補上 v（`kamonu v0.1.0`；tdp S5）。`make build` 不注入，印 `dev`。
  v0.1.2 以前是 `-X main.version=v…`，注入時就帶 v。
- **推 tap 要 PAT**：workflow 內建的 token 寫不了 `homebrew-tap`，要在 repo 的 secret 設 `ACTION_TOKEN`（家族其他 repo 都有）。
  沒設的話用內建的 token，release 會發出去，但推 formula 那一步會失敗。也可以在本機手動發：shell 裡的 `GITHUB_TOKEN` 有權限的話，
  `GOWORK=off goreleaser release --clean`。
- **`make release-check`**：`brews` 在 goreleaser v2 是 deprecated，`goreleaser check` 因此回失敗；family.mk 照樣印出警告，只有這一種
  不算失敗。家族刻意用 `brews`（formula，不是 cask：formula 的執行檔不會被蓋上 `com.apple.quarantine`，沒簽章也能直接跑）。
- 本機試打包：`make snapshot`（不發布、不推 tap），結果在 `dist/`。
- **踩過的坑**：
  - 推 `.github/workflows/` 裡的檔案要 `workflow` 權限（v0.1.0，2026-10-14）。shell 的環境變數 `GITHUB_TOKEN` 只有 `repo`，
    `git push` 會被擋（`refusing to allow a Personal Access Token to create or update workflow … without workflow scope`）。
    `gh` 在 keyring 裡的登入有 `workflow`，所以改 workflow 的 commit 用 `env -u GITHUB_TOKEN git push`。
  - `ACTION_TOKEN` 是用環境變數的 `GITHUB_TOKEN` 設的（user 2026-10-14）：`gh secret set ACTION_TOKEN -R vulcanshen/kamonu`，
    值從 stdin 給，不印出來。它只有 `repo`，推 tap 夠用。
- **發過的版本**：v0.1.0（2026-10-14，tdp v0.6.3，terminu-go v0.1.3），第一版；release、tap 的 formula、`install.sh`／
  `uninstall.sh` 都確認過。
  v0.1.1（2026-10-15，tdp v0.6.9，terminu-go v0.1.8）：匯出時已經有 svg 先問；Draw 的按鈕照表單的動作區移動。
  v0.1.2（2026-10-15，tdp v0.6.11，terminu-go v0.1.10）：Draw 開在 lib 的 popup 疊上（`App.Popup`）；`SAVE AS` 什麼都沒打就按
  `Enter` 等於取消。
