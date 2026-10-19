# kamonu

**Language**: [English](README.md) · 繁體中文

[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**在終端機裡一個字一個字畫 terminu 家族 app 的 mark。**

terminu 家族的每個 app 都有自己那一版的家族 mark（家紋，*kamon*）：一樣的藍色 U 框，框裡用金色寫 app 自己的字，畫在
25 × 19 的格子上。在 kamonu 裡，你先說框裡有哪些字，再用鍵盤把每個字用它自己的顏色畫出來，最後匯出家族用的 svg，
可以直接當 app 的 `docs/icon.svg`：字全是金色，而且照 app 名字的順序一個字一個字標好。

## 特色

- 左邊是你的 mark 清單，右邊預覽 cursor 那一個，畫法跟家族畫 mark 一樣：藍色 U 框裡的金色字。
- 新的 mark 一開始就有 U 框，也可以選全空白。
- 畫的時候每個字一個顏色，一眼看得出哪一格屬於哪個字。
- 用 `hjkl` 與 `Space` 畫；U 框不會被誤改，cursor 會跳過它。
- 名字、字、要不要 U 框，事後都能改，畫好的不會丟。
- 按一個鍵就匯出 svg，每個字照順序標好。
- 可以存回原本的 mark，也可以另存成新的、舊的留著。
- 不會不小心丟掉東西：畫過的沒存就離開會先問，刪除 mark 也會先問。

## 安裝

> kamonu **只支援 macOS / Linux**（WSL 可以），沒有 Windows 版。

**Homebrew**（macOS / Linux）：

```bash
brew install vulcanshen/tap/kamonu
```

**安裝腳本**（最新發布的執行檔裝到 `~/.local/bin`，root 則是 `/usr/local/bin`）：

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/kamonu/main/install.sh | sh
```

從原始碼建置寫在 [`docs/dev-remarks.md`](docs/dev-remarks.md)。

**需求**：支援 24-bit 色（truecolor）的終端機，與 [Nerd Font](https://www.nerdfonts.com/)。

有些字型讓 icon 佔兩格，有些連 powerline 符號（框上膠囊的兩端、cursor 後面的箭頭）也佔兩格。kamonu 啟動時兩種都會自己量；
量錯的話可以自己指定：`KAMONU__ICON_WIDTH` 與 `KAMONU__POWERLINE_WIDTH`（`1` 或 `2`）只給 kamonu，`TERMINU__ICON_WIDTH` 與
`TERMINU__POWERLINE_WIDTH` 給整個 terminu 家族：設一次，家族每個 app 都讀得到。在家族 app 自己的終端機裡跑的程式，外層的 app
會替它把兩個 `TERMINU__` 變數設好。

**第一次啟動時**清單是空的：新增第一個 mark 之前，什麼都不會寫進磁碟。

### 移除

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/kamonu/main/uninstall.sh | sh
```

先移除執行檔，再問你要不要刪掉放 mark 的 `~/.kamonu`，不會自己決定。用 Homebrew 裝的：`brew uninstall kamonu`，不要你的 mark
了，就自己刪掉 `~/.kamonu`。

## 快速開始

1. 執行 `kamonu`。
2. 按 `A`。在 **Name** 上按 `Enter`，打 app 的名字（`filu`）、按 `Enter`；按 `j` 到 **Letters**，按 `Enter`，打框裡的字
   （`fil`：U 是框）、按 `Enter`。按 `Tab` 到 **Create**，按 `Enter`。
3. 用 `hjkl` 移動。在每個字的格子上按那個字的數字：F 按 `1`、I 按 `2`、L 按 `3`。格子下面那一列寫著哪個數字、哪個顏色是哪個字。
4. 按 `Tab` 到按鈕，在 **Save** 上按 `Enter`。
5. 按 `P` 匯出。svg 就是 `~/.kamonu/data/svg/filu.svg`：把它複製成 app 的 `docs/icon.svg`。

## 使用方式

### 清單

| 鍵 | 作用 |
|---|---|
| `Enter` | 畫 cursor 那一個 mark |
| `e` | 改它的名字、字、要不要 U 框 |
| `P` | 匯出它的 svg；已經有了時用 menu 問：**Overwrite** 蓋掉，或 **Export as new** 寫成自己取名的新檔 |
| `X` | 刪除（先問） |
| `A` | 新增 mark：名字、字、要不要從 U 框開始 |
| `Space` | 這裡能做的事 |
| `?` | 這個 panel 的按鍵 |
| `Tab\|1–2` | 另一個 panel |
| `q` | 離開 |

在 **[2] PREVIEW** 上，`Enter` 把那一個 mark 照它在 app 裡的樣子放出來，`e` 打開來畫。終端機太矮、放不下整個 mark 時，用 `j|k`、`u|d`、`gg|G` 捲動。

### 畫

| 鍵 | 作用 |
|---|---|
| `h\|j\|k\|l` | 往左、下、上、右移一格；cursor 會跳過 U 框 |
| `Enter\|Space` | 把這一格換成下一個字；最後一個字之後，變回空的 |
| `Backspace` | 把這一格清成空的，不管它是哪個字 |
| `1`–`6` | 把這一格設成那個字：`1` 是第一個字、`2` 是第二個……；同一個數字再按一次就清空 |
| `p` | 字全變成金色，跟匯出的 svg 一樣；再按一次 `p` 變回分字的顏色 |
| `x` | 把所有格子清空（先問） |
| `Tab` | 在格子與按鈕之間切換 |
| `h\|j\|k\|l` | 在按鈕上：在 **Save As** 與 **Save** 之間移動 |
| `Enter` | 在按鈕上：**Save** 存回原本的 mark；**Save As** 先問新名字，原本的 mark 不動 |
| `?` | 畫格子時的按鍵 |
| `Esc` | 離開；改過的話先問 |

### 改字

每個字照它的位置留著自己的格子。`FIL` 改成 `FAL`，第二個字的格子就是 A 的，形狀還是 I；`KB` 改成 `FIL`，K 的格子變成 F 的、
B 的變成 I 的，L 還沒有格子。字比原本少時，沒了的那幾個字的格子會清空；打開 U 框時，框底下的格子會清空。需要的話再自己畫。

## 設定與資料位置

kamonu 沒有設定檔。所有東西都在 `~/.kamonu/data`；要放在別的地方，設 `KAMONU__DATA`。

| 位置 | 是什麼 |
|---|---|
| `conf/<名字>.yaml` | mark 本身：字、要不要 U 框、每一格，格子寫成一張看得懂的圖 |
| `svg/<名字>.svg` | **匯出**寫出來的檔。已經有了時先問：**Overwrite** 蓋掉它，或 **Export as new** 寫成新的檔，名字自己取；會先帶入 `<名字>_1.svg`（第一個還沒用掉的號碼），直接按 `Enter` 就用它 |

## 限制

- mark 一律是家族的 25 × 19 格、家族的顏色：藍色的 U 框、金色的字。字的顏色只是畫的時候分字用，svg 裡一律是金色。
- 最多 6 個字，每個字是 A 到 Z。
- 改名或刪除 mark 時，已經匯出的 svg 不會跟著改。
- 終端機至少要 56 × 16；比這個小時 kamonu 只告訴你需要多大，拉大就回來。

## 相關連結

- [CHANGELOG](CHANGELOG.md)
- [開發者備忘](docs/dev-remarks.md)

## terminu 家族

kamonu 遵循 [terminu design](https://github.com/vulcanshen/terminu)，terminu 家族的終端機 app 共用的設計。家族的其他成員：

| App | 是什麼 |
|---|---|
| [kbu](https://github.com/vulcanshen/kbu) | Kubernetes TUI dashboard |
| [filu](https://github.com/vulcanshen/filu) | 不用學的 terminal 檔案管理員 |
| [sshu](https://github.com/vulcanshen/sshu) | ssh 與 sftp 的 terminal 前端 |
| [webu](https://github.com/vulcanshen/webu) | 把網頁當文件讀的 terminal 瀏覽器 |
| [locku](https://github.com/vulcanshen/locku) | 帶 PIN 鎖的 terminal 螢幕保護程式 |
| [ccdrawu](https://github.com/vulcanshen/ccdrawu) | Claude Code 的側欄與狀態列 |

## License

[GPL-3.0](LICENSE)
