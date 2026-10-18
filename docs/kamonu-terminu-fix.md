# kamonu 待修：還沒符合 tdp 的地方

照 tdp v0.7.4 對過（2026-10-18，checklist）。每一條：違反了哪一條、在哪裡、現況、該怎麼改。全部修完就刪掉這個檔（tdp E7）。

## 畫格子的 popup 放不下時，下框沒有捲動位置

- **違反**：components/layout/popup 的「內容放不下時」：放不下時下框右邊寫捲動位置，有 cursor 的寫 `N of M`
  （清單 popup-51、popup-53）。
- **在哪裡**：`DRAW: <名字>`，終端機不到 27 列時（例：80 × 24）。`internal/app/editor.go` 的 `Draw`，`TODO(lib KM9)`。
- **現況**：格子已經跟著 cursor 捲，色標、分隔線、按鈕固定在下面；框的寬度已經留了最寬的捲動位置。下框右邊沒寫 cursor 在第幾列：
  lib 的 `PopupBody` 沒有回報捲動位置的方法，框是 lib 畫的。
- **該怎麼改**：等 terminu-go 補上（紀錄 KM9，例：`PopupBody` 多一個可以不實作的 `Scroll() string`），kamonu 回傳
  `<cursor 那一列> of 19`，拿掉 TODO。

## `EXPORT: <名字>` 表單只有一個值

- **違反**：components/dialog/form 的「用途」：好幾個值一起送出才用表單，只有一個值時用 input popup（清單 form-1）。
- **在哪裡**：`[P]` 匯出、svg 已經有了時打開的 `EXPORT: <名字>`（`internal/app/screen.go` 的 `export`）。
- **現況**：一欄 `File`（新檔名），兩顆按鈕：`Overwrite <名字>.svg` 蓋掉原檔（用不到 `File`）、`Export` 寫成 `File`。要選的是兩種
  做法，不只是填一個值；規範沒寫這種情況。
- **該怎麼改**：等定案（紀錄 KM10）：維持表單、規範加例外；或改成帶按鈕的 text input popup；或先開 menu 選做法、再輸入檔名。

## 從 Space menu 刪除時，confirm 底下的 Space menu 不見了

- **違反**：Rules T1 與 F4：target 是短的 confirm 或訊息時保留 source，取消 confirm 回到 source（清單 T1-1、T1-2、F4-1、F4-2）。
- **在哪裡**：`[1]` 的 Space menu → `[X] Delete`（`screen.go` 的 `delete`）；同一類：Space menu → `Draw`、`Edit` 在讀不了的
  kamon 上開出的 error popup。
- **現況**：lib 從 menu 執行動作時先關掉整疊 menu，再跑 `Do`；`Do` 裡開的 confirm 與 error popup 底下就沒有 menu 了，`Esc` 回到
  panel。lib 的 `Action.Ask` 會保留 menu，可是問句是寫死的，帶不了 `Delete <名字>?` 的名字。
- **該怎麼改**：等 terminu-go 補上（紀錄 KM12，例：`Action` 的問句在執行時才產生），Delete 改用它；`Do` 失敗時開的訊息怎麼
  保留 menu，一起定。

## key reference 列出沒有對象的鍵

- **違反**：Rules M6：key reference 裡，對象不存在的鍵不列（清單 M6-6）。
- **在哪裡**：都是 lib 畫的：
  - 清單是空的時，`[1]` 的 key reference 照樣列 `j|k`、`u|d`、`gg|G`。
  - 一列就放得下的 error popup，key reference 照樣列捲動的鍵。
- **現況**：kamonu 自己的 `[2]` 已經只在放不下時列捲動的鍵；上面兩處是 lib 的 `List`、error popup 決定的。
- **該怎麼改**：等 terminu-go 補上（紀錄 KM13）。

## 空狀態的字太長時貼著框、`…` 不是 Overlay0

- **違反**：components/layout/panel 的框（內容離框一格）與 layout/screen 的空狀態（兩句都用 Overlay0）（清單 panel-24、screen-29）。
- **在哪裡**：lib 的 `EmptyState` 把放不下的句子截到整個內寬，`…` 用終端機的預設前景色。kamonu 的兩句提示都縮短了，平常的
  寬度放得下；`[1]` 窄到不到 22 欄、或 `[2]` 窄到不到 32 欄時才會發生。
- **現況**：等 terminu-go 補上（紀錄 KM11）。
- **該怎麼改**：lib 截的時候留左右各一格、`…` 跟著句子的顏色；kamonu 不用改。
