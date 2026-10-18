# kamonu 待修：還沒符合 tdp 的地方

照 tdp v0.7.4 對過（2026-10-18，checklist）。每一條：違反了哪一條、在哪裡、現況、該怎麼改。全部修完就刪掉這個檔（tdp E7）。

## 畫格子的 popup 放不下時，下框沒有捲動位置

- **違反**：components/layout/popup 的「內容放不下時」：放不下時下框右邊寫捲動位置，有 cursor 的寫 `N of M`
  （清單 popup-51、popup-53）。
- **在哪裡**：`DRAW: <名字>`，終端機不到 27 列時（例：80 × 24）。`internal/app/editor.go` 的 `Draw`，`TODO(lib KM9)`。
- **現況**：格子已經跟著 cursor 捲，色標、分隔線、按鈕固定在下面；下框右邊沒寫 cursor 在第幾列。lib 的 `PopupBody` 沒有回報
  捲動位置的方法，框是 lib 畫的。
- **該怎麼改**：等 terminu-go 補上（紀錄 KM9，例：`PopupBody` 多一個可以不實作的 `Scroll() string`），kamonu 回傳
  `<cursor 那一列> of 19`，拿掉 TODO。

## `EXPORT: <名字>` 表單只有一個值

- **違反**：components/dialog/form 的「用途」：好幾個值一起送出才用表單，只有一個值時用 input popup（清單 form-1）。
- **在哪裡**：`[P]` 匯出、svg 已經有了時打開的 `EXPORT: <名字>`（`internal/app/screen.go` 的 `export`）。
- **現況**：一欄 `File`（新檔名），兩顆按鈕：`Overwrite <名字>.svg` 蓋掉原檔（用不到 `File`）、`Export` 寫成 `File`。要選的是兩種
  做法，不只是填一個值；規範沒寫這種情況。
- **該怎麼改**：等定案（紀錄 KM10）：維持表單、規範加例外；或改成帶按鈕的 text input popup；或先開 menu 選做法、再輸入檔名。
