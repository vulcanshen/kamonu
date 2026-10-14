# kamonu

terminu 家族的一員：畫家族 mark（家紋）的 TUI app。名字是日文的「家紋」（kamon）加上家族的 `u`。

- module：`github.com/vulcanshen/kamonu`
- 授權：GPL-3.0（跟家族其他 app 一樣）
- 用 Go，建在 **terminu-go** 上（`github.com/vulcanshen/terminu-go`，Go 套件名 `terminu`；本機在 `~/Documents/sideproj/terminu-go`）。
  2026-10-14 推上 GitHub，第一個 tag 是 v0.1.0（實作 tdp v0.6.0）：
  - `go.mod` 寫 terminu-go 打過的 tag，不用 `replace`，也不指向 commit 的 pseudo-version。
  - 開發時用不進版控的 `go.work` 把本機的 `../terminu-go` 接進來（`go.work`、`go.work.sum` 加進 `.gitignore`）。
  - 發版前、commit 前用 `GOWORK=off go build ./...`、`GOWORK=off go test ./...` 確認拿 GitHub 版本也過得了；過不了表示用到了
    lib 還沒發版的東西，要等 terminu-go 打新 tag。
  - 切換的步驟 terminu-go 的 session 已經直接告訴你。
- 規範：`~/Documents/sideproj/terminu`（tdp），照**最新 tag** 做（現在是 **tdp v0.6.5**）。
- **kamonu 是第一個真正用 terminu-go 寫的 app**：terminu-go 的展示場證明了 lib 照規範畫得出來，kamonu 證明真的 app 用得順。

上面這些都已經跟 user 定案，不要重提。

## 它做什麼（已定）

2026-10-13 在 terminu 的 session 跟 user 定的：

- **畫面**：
  - 左邊 panel `[1] KAMON`：icon 檔案清單，資料目錄 `~/.kamonu/data/` 裡的 `*.svg`（例：`locku.svg`、`webu.svg`）。
  - 右邊 panel：cursor 那一筆的預覽。
- **動作**：
  - `[Enter] Edit`：左邊清單的 item operation（原本是 `[E]dit`，2026-10-14 改成 `Enter`，見 `docs/dev-remarks.md`）。
  - `[A]dd`：panel operation。
  - Edit 打開編輯 popup；Add 先開 `NEW KAMON` 表單（名字、要不要注入 U 框），`Create` 當場建檔，再打開編輯 popup
    （2026-10-14 改，見 `docs/dev-remarks.md`）。
- **編輯 popup**：
  - 外框、標題、hint 照 tdp。內容是格子，每一格當成 checkbox，`Space` 切換亮暗（rules K5 的 checkbox 例外）。
  - 下面是動作列，`Tab` 在欄位區與動作列之間切換（跟 tdp 的表單一樣）。
  - 按鈕是 `Save`（存回原檔）與 `Save As`（存成新檔）。
- **格子的畫法、svg 的格式照 rules S5**：
  - 一格一個 `nf-fa-square`、佔兩欄；底是 Surface0；U 框 `#205090`；金色 `#f2b753`。
  - viewBox 250×250，一格 10×10 的 rect。
  - XML 註解分字。kamonu **不分字**：存好之後，由開發 app 的人自己加註解。
- **位置與名字**：資料 `~/.kamonu/data/`（rules E3）；環境變數 `KAMONU__…`（E2）。

## user 提了、還沒細定的（跟 user 一題一題定）

- ~~U 底盤是一個檔案、建立時選空白或帶 U、建立時選大小~~：2026-10-14 已定，見 `docs/dev-remarks.md`（U 框是獨立的檔、不在清單裡；
  建立時選要不要注入；**只有家族大小 25×19**，full size 與自訂大小已否決 —— kamonu 是家族家徽的編輯器）。
- **當 locku 的保護程式**：locku 讀 kamonu 畫的圖當一種 saver。這是 locku 的功能，locku 改用 lib 之後再談，這裡先不做。

這些是 app 自己的設計（tdp 的「由 app 決定」與 concept zone），在這個 session 跟 user 定，定了寫進 `docs/dev-remarks.md`。

## 規範是唯一的依據

照 terminu-go 的做法：

- **讀什麼**：
  - terminu 最新 tag 的 `principle/README-zh_TW.md`、`principle/rules-zh_TW.md`、`components/` 全部。
  - 繁中是原文，英文是翻譯。
- **不讀什麼**：
  - 五個 app 的程式（kbu、filu、sshu、webu、locku）。
  - terminu 的 `.local/`、`.checkpoints/`、`vtp/`。唯一要碰的是你自己的盲點紀錄。
- **terminu-go**：當 lib 用。可以讀它的 API、文件與範例（`examples/showcase`），**不改它**。

**盲點**：碰到規範沒寫、模糊、矛盾、方案不夠，或照做出來不好用的地方，不在程式裡自己決定，記下來。判斷方式跟 terminu-go 一樣：
使用者看得到或按得到的，聽規範的；app 自己的領域（concept zone、規範寫「由 app 決定」的），跟 user 定。

**lib 不夠用**：
- 判斷方式：terminu-go 少了 kamonu 要的東西（例：popup 裡放 app 自己畫的格子、動作列兩個按鈕），先判斷是哪一種：
  - 規範沒寫：記成盲點。
  - 規範寫了、lib 沒做：記成「lib 要補」。
- 處理方式：兩種都不自己改 terminu-go，跟 user 說，由 terminu 的 session 協調。
- 等的時候：可以先在 kamonu 裡暫時自己畫，標 `TODO(lib KM<n>)`。

### 記在哪、怎麼記

記在 `~/Documents/sideproj/terminu/.local/family-fix/kamonu/README.md`（下面叫「紀錄」）。編號用 **KM**，跟 terminu-go 紀錄的 G 分開：

```
### KM<n> <一句話標題>
- 狀態：待定
- 種類：盲點（規範）／lib 要補
- 規範：<檔案> 的 <段落或規則編號>（tdp v0.6.5）
- 卡在：kamonu 需要知道什麼才能做下去
- 可能：A …；B …（各附一句後果）
- 暫時：跳過／暫時照 A 做（程式裡標 `TODO(tdp KM<n>)` 或 `TODO(lib KM<n>)`）
```

寫紀錄前先重讀檔案（terminu 的 session 也會寫它）。新版 tdp 發布後，照 terminu-go 的做法跟上：狀態改「已補（tdp vX.Y.Z）」，用
`git -C ~/Documents/sideproj/terminu diff <舊 tag> <新 tag> -- principle components` 看完整的改動。

## repo 一開始要做的

1. `git init`、`go.mod`、`LICENSE`（GPL-3.0 全文，可以從 terminu-go 複製）。
2. `README.md` 與 `README-zh_TW.md`（照 rules E7 與 user 的 README 規範：README 是工具介紹，開發者的東西放 `docs/dev-remarks.md`）。
3. `CHANGELOG.md`。
4. 自己的家族 mark：用 kamonu 自己畫（U 框裡 K、A、M、O、N），存成 `docs/icon.svg`，再照 S5 加分字註解。
5. 建 GitHub repo、push、打 tag、發布都先問 user。

## 參考

`.local/reference/iconedit/` 是同一件事的第一版（原本在 terminu 的 `tools/`，user 要搬過來、terminu 那邊刪掉）。user 試過：
「使用體驗上好糟糕，但是功能是 OK 的」。它的 `mark.go`（U 框的格子、svg 的寫法）可以參考；它不照 tdp，畫面與按鍵不要照抄。
`.local/` 不進版控：`git init` 時在 `.gitignore` 加上 `.local/`。

## 跟 user 合作的方式

- 一次問一題，用問句寫明要決定什麼，列 A／B 選項附上建議與理由。不要寫一串直述句再問「可以嗎」。
- 先講操作（使用者會按什麼、看到什麼），再寫程式；看得到的東西先做終端機樣稿。
- 提案前先查規範怎麼說。
- user 看不懂時，用一個具體的例子重講。
- 只改這個 repo 與自己的紀錄；不改 terminu 的規範、terminu-go、五個 app。
- 一律用繁體中文對話；commit 不加任何 Claude 署名。
