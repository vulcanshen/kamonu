# kamonu 待修：還沒符合 tdp 的地方

照 tdp v0.9.1 對過（2026-10-20）。每一條：違反了哪一條、在哪裡、現況、該怎麼改。全部修完就刪掉這個檔（tdp E7）。

這裡的每一條都等 terminu-go：kamonu 能自己做的都做了。lib 發版後 kamonu 升級一次，照下面「該怎麼改」做完，拿掉程式裡對應的
`TODO(lib KM…)`。

## 讀不了的 mark，問句寫「建一個新的」

- **違反**：Rules E3 的「讀不了時」：資料是一筆一個檔時，讀不了的那一筆按 `Enter` 只改名成 `.bak`、不另外建，其他照常讀；問句寫
  `Move it to <檔名>.bak and go on without it?`（清單 E3-22，tdp v0.9.1）。
- **在哪裡**：啟動時有 mark 讀不了，開出的 `MOVE ASIDE`（`screen.go` 的 `start`，`TODO(lib KM18)`）。
- **現況**：做的事已經對了：`Enter` 只把那一個改名成 `<名字>.yaml.bak`，不另外建，再檢查下一個，其他照常讀。問句是 lib 的
  `App.MoveAside` 寫的，terminu-go v0.4.0 一律寫 `Move it to <名字>.yaml.bak and start a new one?`。
- **該怎麼改**：lib 加上「一筆一個檔」的寫法後（紀錄 KM18），讀不了的 mark 改用它；`conf/` 目錄本身讀不了時照舊。
