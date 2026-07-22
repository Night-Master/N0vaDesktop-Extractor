### **这东西干嘛用的**
人工桌面的图片和视频都是ndf格式，无法直接查看，其中视频更是需要源文件去掉前两个字节才能正常播放，因此我写了这个工具来帮助大家提取人工桌面的图片和视频为png/jpg和mp4

### 特性
- 打开网页自动扫描所有盘符的常见安装路径（`Program Files`、`Program Files (x86)`、盘符根目录下的 `N0vaDesktop\N0vaDesktopCache\game`），点击扫描结果即可填入，只扫到一个时自动填好
- 输出文件按内容 MD5 命名，重复点击转换不会产生重复文件
- 自动过滤宽高相乘小于 936000 像素的图片和视频
- 前端已用 go embed 内嵌进 exe，单文件即可运行
- 提供 Windows amd64 / 386 / arm64 三种平台的 exe，配套 bat 一键启动服务并打开浏览器

### 怎么用
1. 下载对应平台的 exe 和同名 bat（`dist` 目录，或 Releases）
2. 双击 bat，它会自动启动服务并打开浏览器；也可以先双击 exe，再手动访问 http://127.0.0.1:8080
3. 网页已自动扫描安装路径，确认填入的源目录（通常是 `...\N0vaDesktopCache\game`），也可以手动输入其他目录
4. 点击「转换文件」，提取结果在 exe 同目录的 `output` 文件夹里
5. 不用了关掉 exe 的控制台窗口即可

### 自行编译
```bash
go build -o N0vaDesktop-Extractor.exe .
# 交叉编译其他平台
GOOS=windows GOARCH=386 go build -o N0vaDesktop-Extractor-386.exe .
GOOS=windows GOARCH=arm64 go build -o N0vaDesktop-Extractor-arm64.exe .
```

### What is this for?
The images and videos of Artificial Desktop (N0vaDesktop) are in ndf format and cannot be viewed directly. Videos, in particular, require the first two bytes of the source file to be removed to play normally. This tool extracts them as png/jpg and mp4.

### Features
- Automatically scans the common install paths on every drive (`Program Files`, `Program Files (x86)`, and `N0vaDesktop\N0vaDesktopCache\game` at drive roots) when the page opens; click a result to fill it in, and a single result is filled automatically
- Output files are named by content MD5, so repeated conversions never create duplicates
- Images and videos whose width × height is below 936,000 pixels are skipped
- Frontend is embedded into the exe via go embed — a single file is all you need
- Prebuilt exes for Windows amd64 / 386 / arm64, each with a bat that starts the service and opens the browser

### How to use
1. Download the exe and matching bat for your platform (`dist` directory or Releases)
2. Double-click the bat — it starts the service and opens the browser. Or run the exe first and visit http://127.0.0.1:8080 manually
3. The page has already scanned for install paths; confirm the detected source directory (usually `...\N0vaDesktopCache\game`), or type another one
4. Click "转换文件" (Convert); extracted files are in the `output` folder next to the exe
5. Close the exe console window to stop the service

### Build from source
```bash
go build -o N0vaDesktop-Extractor.exe .
```

### これは何に使うのですか？
人工デスクトップ（N0vaDesktop）の画像と動画は ndf 形式で直接閲覧できません。特に動画は先頭2バイトを削除しないと再生できないため、png/jpg と mp4 として抽出するこのツールを作りました。

### 特徴
- ページを開くと全ドライブの一般的なインストールパス（`Program Files`、`Program Files (x86)`、ルート直下の `N0vaDesktop\N0vaDesktopCache\game`）を自動スキャンし、クリックで入力できます
- 出力ファイルは内容の MD5 で命名されるため、繰り返し変換しても重複しません
- 幅×高さが 936,000 ピクセル未満の画像・動画はスキップします
- go embed でフロントエンドを exe に内蔵、単一ファイルで動作
- Windows amd64 / 386 / arm64 向けの exe と、サービス起動＋ブラウザを開く bat を同梱

### 使い方
1. プラットフォームに合った exe と bat をダウンロード（`dist` または Releases）
2. bat をダブルクリックするとサービスが起動しブラウザが開きます（exe を先に起動して http://127.0.0.1:8080 に手動アクセスでも可）
3. スキャン結果のソースディレクトリを確認（通常は `...\N0vaDesktopCache\game`）、手動入力も可能
4. 「转换文件」をクリック。結果は exe と同じ場所の `output` フォルダに保存されます
5. 終了するには exe のコンソールを閉じてください

### Wofür ist das?
Die Bilder und Videos von Artificial Desktop (N0vaDesktop) sind im ndf-Format und können nicht direkt angezeigt werden. Videos müssen zusätzlich um die ersten zwei Bytes gekürzt werden. Dieses Tool extrahiert sie als png/jpg und mp4.

### Funktionen
- Beim Öffnen der Seite werden die üblichen Installationspfade aller Laufwerke (`Program Files`, `Program Files (x86)`, `N0vaDesktop\N0vaDesktopCache\game` im Laufwerksstamm) automatisch gescannt; ein Klick übernimmt den Pfad
- Ausgabedateien werden nach Inhalts-MD5 benannt — wiederholte Konvertierung erzeugt keine Duplikate
- Bilder und Videos mit Breite × Höhe unter 936.000 Pixeln werden übersprungen
- Frontend per go embed in die exe eingebettet — eine einzige Datei genügt
- Fertige exes für Windows amd64 / 386 / arm64, jeweils mit bat zum Starten des Dienstes und Öffnen des Browsers

### Wie es benutzt wird
1. Lade exe und passende bat für deine Plattform herunter (`dist` oder Releases)
2. Doppelklicke die bat — sie startet den Dienst und öffnet den Browser (oder exe starten und http://127.0.0.1:8080 manuell öffnen)
3. Der gescannte Quellordner ist bereits eingetragen (meist `...\N0vaDesktopCache\game`), alternativ manuell eingeben
4. Auf „转换文件" klicken; die extrahierten Dateien liegen im Ordner `output` neben der exe
5. Zum Beenden das Konsolenfenster der exe schließen

### À quoi cela sert-il ?
Les images et vidéos d'Artificial Desktop (N0vaDesktop) sont au format ndf et ne peuvent pas être visualisées directement. Les vidéos nécessitent en outre la suppression des deux premiers octets. Cet outil les extrait en png/jpg et mp4.

### Fonctionnalités
- À l'ouverture de la page, les chemins d'installation courants de tous les lecteurs (`Program Files`, `Program Files (x86)`, `N0vaDesktop\N0vaDesktopCache\game` à la racine) sont scannés automatiquement ; un clic remplit le champ
- Les fichiers de sortie sont nommés par MD5 du contenu : aucune duplication en cas de conversions répétées
- Les images et vidéos dont largeur × hauteur est inférieure à 936 000 pixels sont ignorées
- Frontend intégré à l'exe via go embed — un seul fichier suffit
- Exes pour Windows amd64 / 386 / arm64, avec un bat qui lance le service et ouvre le navigateur

### Comment l'utiliser
1. Téléchargez l'exe et le bat correspondant à votre plateforme (`dist` ou Releases)
2. Double-cliquez sur le bat — il démarre le service et ouvre le navigateur (ou lancez l'exe puis ouvrez http://127.0.0.1:8080)
3. Le dossier source détecté est déjà rempli (généralement `...\N0vaDesktopCache\game`), ou saisissez-le manuellement
4. Cliquez sur « 转换文件 » ; les fichiers extraits sont dans le dossier `output` à côté de l'exe
5. Fermez la console de l'exe pour arrêter le service
