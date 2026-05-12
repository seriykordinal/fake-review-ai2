# =============================================================================
# start.ps1 — Запуск проекта FakeCheck (Windows)
# Запуск: .\start.ps1
# =============================================================================

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot

function Step($msg)  { Write-Host "`n>>> $msg" -ForegroundColor Cyan }
function OK($msg)    { Write-Host "    [OK] $msg" -ForegroundColor Green }
function Warn($msg)  { Write-Host "    [WARN] $msg" -ForegroundColor Yellow }
function Fail($msg)  { Write-Host "    [ERROR] $msg" -ForegroundColor Red; exit 1 }

# ── 1. Компиляция TypeScript ──────────────────────────────────────────────────
Step "Компиляция TypeScript..."
Set-Location $Root
try {
    npx tsc --project tsconfig.json 2>&1 | Out-Null
    OK "TypeScript скомпилирован"
} catch {
    Fail "Ошибка компиляции TypeScript: $_"
}

# ── 2. Go-сервер ──────────────────────────────────────────────────────────────
Step "Запуск Go-сервера..."
Set-Location "$Root\backend-go"
$goProc = Start-Process "go" -ArgumentList "run", "." -PassThru -NoNewWindow
OK "Go-сервер запущен (PID $($goProc.Id)) → http://localhost:8080"
Start-Sleep -Seconds 2

# ── 3. Проверка ML-модели ─────────────────────────────────────────────────────
Step "Проверка ML-модели..."
Set-Location $Root
$model  = "backend-python\models\fake_review_model.h5"
$tfidf  = "backend-python\models\tfidf_vectorizer.pkl"
$ds     = "backend-python\dataset\wb_reviews.csv"

if (-not (Test-Path $model) -or -not (Test-Path $tfidf)) {
    Warn "Модель не найдена — запускаем обучение"

    if (-not (Test-Path $ds)) {
        Warn "Датасет не найден — скачиваем"
        try {
            python backend-python\dataset\get_dataset.py
            OK "Датасет скачан"
        } catch {
            Warn "Не удалось скачать датасет: $_"
            Warn "Запустите вручную: python backend-python\dataset\get_dataset.py"
        }
    } else {
        OK "Датасет уже есть"
    }

    try {
        python backend-python\train.py
        OK "Модель обучена"
    } catch {
        Warn "Ошибка обучения: $_"
        Warn "Запустите вручную: python backend-python\train.py"
    }
} else {
    OK "Модель уже обучена"
}

# ── 4. Python-сервер ──────────────────────────────────────────────────────────
Step "Запуск Python-сервера..."
$pyProc = Start-Process "python" `
    -ArgumentList "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000" `
    -WorkingDirectory "$Root\backend-python" `
    -PassThru -NoNewWindow
OK "Python-сервер запущен (PID $($pyProc.Id)) → http://localhost:8000"

# ── Итог ──────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "=============================================" -ForegroundColor White
Write-Host "  Проект запущен!" -ForegroundColor Green
Write-Host "  Сайт:   http://localhost:8080" -ForegroundColor White
Write-Host "  ML API: http://localhost:8000/health" -ForegroundColor White
Write-Host "  Для остановки нажмите Ctrl+C" -ForegroundColor Gray
Write-Host "=============================================" -ForegroundColor White

try {
    $goProc.WaitForExit()
} finally {
    if (-not $pyProc.HasExited) { $pyProc.Kill() }
}
