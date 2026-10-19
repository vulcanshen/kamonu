# kamonu 待修：還沒符合 tdp 的地方

照 tdp v0.8.0 對過（2026-10-19）。每一條：違反了哪一條、在哪裡、現況、該怎麼改。全部修完就刪掉這個檔（tdp E7）。

這裡的每一條都等 terminu-go 的第三批（定案見 terminu 的紀錄）：kamonu 能自己做的都做了，剩下的是 v0.1.13 的 lib 畫不出來的。
lib 發版後 kamonu 升級一次，照下面「該怎麼改」做完，拿掉程式裡對應的 `TODO(lib KM…)`。

## 畫格子的 popup 放不下時，下框沒有捲動位置

- **違反**：components/layout/popup 的「內容放不下時」：放不下時下框右邊寫捲動位置，有 cursor 的寫 `N of M`（清單 popup-51、popup-53）。
- **在哪裡**：`DRAW: <名字>`，終端機不到 27 列時（例：80 × 24）。`internal/app/editor.go` 的 `Draw`，`TODO(lib KM9)`。
- **現況**：格子跟著 cursor 捲，色標、分隔線、按鈕固定在下面；框的寬度已經留了最寬的捲動位置。下框右邊寫不出來：lib 的
  `PopupBody` 沒有回報捲動位置的方法。
- **該怎麼改**：lib 補上後（紀錄 KM9），回傳 `<cursor 那一列> of 19`。

## 從 menu 開出的東西，取消或失敗後沒有回到 menu

- **違反**：tdp T1 與 tdp F4（tdp v0.8.0 照「確定之後的動作」分：取消一律回到 menu；失敗開的 error popup 關掉回到 menu，
  components/layout/popup 的錯誤列）（清單 T1、F4 的條目、popup-69）。
- **在哪裡**：
  - Space menu → `Draw`、`Edit` 在讀不了的 kamon 上開出的 error popup（`screen.go` 的 `draw`、`edit`）。
  - 匯出的 menu → `Export as new…` 開的輸入框上按 `Esc`（`screen.go` 的 `export`）。
- **現況**：lib 從 menu 執行動作時先關掉整疊 menu，再跑動作；動作開出來的東西底下就沒有 menu 了。`Delete` 走 lib 的
  `Action.Ask`，取消回到 menu、確定後清掉，已經對了（問句靠 `Hidden` 寫成目前那一筆，`TODO(lib KM12)`，等 `AskFor`）。
- **該怎麼改**：lib 讓動作標「做完之後保留或清掉 source」後（紀錄 KM12），Draw、Edit 的失敗與 `Export as new…` 的取消回到 menu。

## 新檔名沒有用建議值帶入

- **違反**：components/input/README 的建議值：沒有舊值、app 有合理的預設名時，用灰字提議，不碰按 `Enter` 就用它（tdp v0.8.0）。
- **在哪裡**：匯出的 `Export as new…` 開的 `EXPORT AS`（`screen.go` 的 `export`，`TODO(lib KM10)`）。
- **現況**：預設的 `<名字>_<n>.svg` 寫在說明列（`e.g. …`），要自己打；lib 的 `TextInput` 還沒有建議值。
- **該怎麼改**：lib 加上建議值後，把 `<名字>_<n>.svg` 給它，說明列拿掉 `e.g.`。

## 匯出的 menu 標題沒有寫名字

- **違反**：components/layout/popup 的外框：要寫對哪一個做時，名稱接在標題後面、照原樣（清單 popup-6）。
- **在哪裡**：匯出的 menu（`screen.go` 的 `export`，`TODO(lib KM10)`）。
- **現況**：terminu-go v0.1.13 的 `Menu` 把整串標題轉大寫，寫進名字會被改大小寫，暫時只寫 `EXPORT`；列上寫了
  `Overwrite <名字>.svg`。
- **該怎麼改**：改用 lib 帶名字的 menu（照原樣的 subject），標題 `EXPORT: <名字>`。

## 成功的 toast 不是綠色

- **違反**：components/dialog/toast 的 Success：成功做完的訊息用 Success，標題與文字 Green（tdp v0.8.0；清單 color-12）。
- **在哪裡**：匯出成功的 `Exported to …`（`screen.go` 的 `exported`，`TODO(lib KM15)`）。
- **現況**：terminu-go v0.1.13 的 toast 只有 Info、Warning、Error，暫時用 Info。
- **該怎麼改**：lib 有 Success 後改用它。

## 熱鍵沒有原地包進 label

- **違反**：tdp M5（v0.8.0）：熱鍵的字母出現在 label 裡就原地包，不分大小寫，括號裡照實際的鍵（例：`Ex[P]ort`、`[e]dit`）。
- **在哪裡**：`[1]` 的 `Export`（`P`）畫成 `[P] Export`；`[2]` 的 `Edit`（`e`）畫成 `[e] Edit`。
- **現況**：label 已經是正常的大小寫；terminu-go v0.1.13 比對時大小寫算數，找不到就放前面。
- **該怎麼改**：lib 改成不分大小寫比對後自然就對了（紀錄 KM16），kamonu 不用改。

## key reference 列出沒有對象的鍵

- **違反**：tdp M6（v0.8.0）：清單空了、內容放得下算對象不存在，不列移動、捲動鍵（清單 M6-6）。
- **在哪裡**：一列就放得下的 error popup，key reference 照樣列捲動的鍵（lib 畫的）。
- **現況**：kamonu 自己的 `[2]` 只在放不下時列；`[1]` 清單是空的時不列移動鍵（包一層 lib 的 `List`，`TODO(lib KM13)`）。
- **該怎麼改**：lib 的 MoveKeys 依當下狀態給之後，error popup 自然就對了；kamonu 拿掉包的那一層。

## 空狀態的字太長時被截到貼著框

- **違反**：components/layout/screen 的空狀態（tdp v0.8.0：太長在空白處折行，左右照樣留一格）（清單 panel-24、screen-29）。
- **在哪裡**：lib 的 `EmptyState`。kamonu 的提示平常的寬度放得下；`[1]` 內寬不到 22 欄、`[2]` 內寬不到 32 欄時才會截。
- **現況**：等 lib（紀錄 KM11）。
- **該怎麼改**：lib 改成折行後自然就對了，kamonu 不用改。
