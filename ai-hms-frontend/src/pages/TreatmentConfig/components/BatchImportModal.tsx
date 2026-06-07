// 批量导入模态框 - Excel 版本

import { useState, useCallback, useRef, memo, type ChangeEvent, type MouseEvent } from 'react'
import { X, Upload, CheckCircle2, XCircle, AlertCircle, Loader2, FileSpreadsheet, Download } from 'lucide-react'
import * as XLSX from 'xlsx'
import { getToken } from '@/utils/token'

// 导入状态类型
type ImportStatus = 'idle' | 'validating' | 'ready' | 'importing' | 'success' | 'error' | 'partial' | 'confirm_categories'

// 导入结果类型
interface ImportResult {
  success: number
  failed: number
  skipped?: number
  errors: string[]
  needConfirmCategories?: boolean
  missingCategories?: string[]
  data?: unknown[] // 保存原始数据用于确认后重新导入
}

interface ExcelColumn {
  key: string
  label: string
  required: boolean
  example?: string
}

interface Props {
  title: string
  columns: ExcelColumn[]
  onImport: (
    data: unknown[],
    autoAddCategories?: boolean,
    onProgress?: (percent: number) => void
  ) => Promise<ImportResult>
  onClose: () => void
  onExport?: () => Promise<void>
  maxSize?: number // MB
}

interface ValidationError {
  row: number
  field: string
  message: string
}

export const BatchImportModal = memo(function BatchImportModal({
  title,
  columns,
  onImport,
  onClose,
  onExport,
  maxSize = 5
}: Props) {
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const [status, setStatus] = useState<ImportStatus>('idle')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [previewData, setPreviewData] = useState<unknown[] | null>(null)
  const [totalDataCount, setTotalDataCount] = useState(0)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [progress, setProgress] = useState(0)
  // 新增：保存缺失的分类和原始数据
  const [missingCategories, setMissingCategories] = useState<string[]>([])
  const [pendingData, setPendingData] = useState<unknown[] | null>(null)

  // 解析 Excel 文件（必须在 handleFileChange 之前声明）
  const parseExcelFile = useCallback(async (file: File): Promise<unknown[]> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()

      reader.onload = (e) => {
        try {
          const data = e.target?.result
          // 使用 ArrayBuffer 类型
          const workbook = XLSX.read(data, { type: 'array' })

          // 读取第一个工作表
          const firstSheetName = workbook.SheetNames[0]
          const worksheet = workbook.Sheets[firstSheetName]

          // 转换为 JSON
          const jsonData = XLSX.utils.sheet_to_json(worksheet, {
            header: 1,
            defval: ''
          })

          // 获取表头
          const headers = jsonData[0] as string[]

          // 映射列名
          const mappedData = jsonData.slice(1).map((row: unknown) => {
            const rowData = Array.isArray(row) ? row : []
            const item: Record<string, unknown> = {}
            headers.forEach((header, index) => {
              const col = columns.find(c => c.label === header)
              if (col) {
                // 确保 undefined 和 null 都转换为空字符串
                item[col.key] = (rowData[index] === undefined || rowData[index] === null) ? '' : rowData[index]
              }
            })
            return item
          }).filter(item => Object.keys(item).length > 0) // 过滤空行

          resolve(mappedData)
        } catch (err) {
          reject(err)
        }
      }

      reader.onerror = () => reject(new Error('文件读取失败'))
      // 使用 readAsArrayBuffer 替代已弃用的 readAsBinaryString
      reader.readAsArrayBuffer(file)
    })
  }, [columns])

  // 生成 Excel 模板
  const generateTemplate = useCallback(() => {
    // 创建表头行（带必填标记）
    const headers = columns.map(col => col.label)

    // 创建示例数据行（必填项用 * 标记）
    const exampleRow: Record<string, string> = {}
    columns.forEach(col => {
      const reqMark = col.required ? '*' : ''
      exampleRow[col.label] = `${reqMark}${col.example || ''}`
    })

    const templateData = [exampleRow]

    const ws = XLSX.utils.json_to_sheet(templateData, { header: headers })
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, '导入模板')

    // 设置列宽
    ws['!cols'] = columns.map(() => ({ wch: 18 }))

    // 下载文件
    XLSX.writeFile(wb, '导入模板.xlsx')
  }, [columns])

  // 处理文件选择
  const handleFileChange = useCallback(async (e: ChangeEvent<HTMLInputElement>) => {
    const inputEl = e.target
    const file = inputEl.files?.[0]
    if (!file) {
      inputEl.value = ''
      return
    }

    try {
      // 检查文件大小
      if (file.size > maxSize * 1024 * 1024) {
        setStatus('error')
        setValidationErrors([{ row: 0, field: 'file', message: `文件大小超过 ${maxSize}MB` }])
        return
      }

      // 检查文件扩展名
      const fileName = file.name.toLowerCase()
      if (!fileName.endsWith('.xlsx') && !fileName.endsWith('.xls')) {
        setStatus('error')
        setValidationErrors([{ row: 0, field: 'file', message: '请选择 Excel 格式文件 (.xlsx 或 .xls)' }])
        return
      }

      setSelectedFile(file)
      setStatus('validating')
      setValidationErrors([])
      setPreviewData(null)

      const data = await parseExcelFile(file)

      if (data.length === 0) {
        throw new Error('文件中没有数据')
      }

      // 验证数据格式
      const errors: ValidationError[] = []
      data.forEach((item, index) => {
        const row = item as Record<string, unknown>
        columns.forEach(col => {
          if (col.required && !row[col.key]) {
            errors.push({
              row: index + 2, // Excel 行号从1开始，加上表头行
              field: col.label,
              message: `缺少必填字段 "${col.label}"`
            })
          }
        })
      })

      if (errors.length > 0) {
        setValidationErrors(errors)
        setStatus('error')
      } else {
        setPreviewData(data.slice(0, 3)) // 只显示前3条预览
        setTotalDataCount(data.length) // 保存总数据量
        setStatus('ready')
      }
    } catch (err) {
      setValidationErrors([{ row: 0, field: 'excel', message: err instanceof Error ? err.message : '文件格式错误' }])
      setStatus('error')
    } finally {
      // 允许重复选择同一个文件
      inputEl.value = ''
    }
  }, [maxSize, columns, parseExcelFile])

  const handleSelectFile = useCallback((e: MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    fileInputRef.current?.click()
  }, [])

  // 执行导入
  const handleImport = useCallback(async () => {
    if (!selectedFile || status !== 'ready') return

    // 检查登录状态
    const token = getToken()
    if (!token) {
      setValidationErrors([{ row: 0, field: 'auth', message: '用户未登录，请先登录后再试' }])
      setStatus('error')
      return
    }

    setStatus('importing')
    setProgress(0)

    try {
      const data = await parseExcelFile(selectedFile)
      const result = await onImport(data, false, setProgress)

      // 检查是否需要确认添加分类
      if (result.needConfirmCategories && result.missingCategories) {
        setMissingCategories(result.missingCategories)
        setPendingData(result.data || data)
        setStatus('confirm_categories')
        return
      }

      setImportResult(result)
      setStatus(result.failed === 0 && (result.skipped ?? 0) === 0 ? 'success' : 'partial')
    } catch (err) {
      setImportResult({ success: 0, failed: 1, errors: [err instanceof Error ? err.message : '导入失败'] })
      setStatus('error')
    }
  }, [selectedFile, status, onImport, parseExcelFile])

  // 确认添加缺失的分类并继续导入
  const handleConfirmAddCategories = useCallback(async () => {
    if (!pendingData) return

    setStatus('importing')
    setProgress(0)

    try {
      // 调用 onImport 并传入 autoAddCategories=true
      const result = await onImport(pendingData, true, setProgress)

      setImportResult(result)
      setStatus(result.failed === 0 && (result.skipped ?? 0) === 0 ? 'success' : 'partial')
    } catch (err) {
      setImportResult({ success: 0, failed: 1, errors: [err instanceof Error ? err.message : '导入失败'] })
      setStatus('error')
    }
  }, [pendingData, onImport])

  // 导出数据
  const handleExport = useCallback(async () => {
    if (onExport) {
      await onExport()
    }
  }, [onExport])

  // 重置状态
  const handleReset = useCallback(() => {
    setSelectedFile(null)
    setValidationErrors([])
    setPreviewData(null)
    setTotalDataCount(0)
    setImportResult(null)
    setProgress(0)
    setMissingCategories([])
    setPendingData(null)
    setStatus('idle')
  }, [])

  // 关闭弹窗
  const handleClose = useCallback(() => {
    if (status === 'importing') return
    onClose()
  }, [status, onClose])

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={handleClose}>
      <div className="bg-white rounded-3xl p-8 w-[600px] max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <input
          ref={fileInputRef}
          type="file"
          accept=".xlsx,.xls"
          onChange={handleFileChange}
          className="absolute w-0 h-0 opacity-0 pointer-events-none"
          tabIndex={-1}
          aria-hidden="true"
        />

        {/* 标题栏 */}
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-black text-slate-800">{title}</h3>
          <button
            onClick={handleClose}
            disabled={status === 'importing'}
            className={`p-1.5 rounded-lg transition-all ${
              status === 'importing' ? "text-slate-300 cursor-not-allowed" : "hover:bg-slate-100 text-slate-400"
            }`}
          >
            <X size={18} />
          </button>
        </div>

        {/* 说明文字 */}
        <div className="space-y-4">
          <p className="text-sm text-slate-600">
            请选择 Excel 格式的文件（.xlsx 或 .xls，最大 {maxSize}MB）
          </p>

          {/* 必填字段说明 */}
          <div className="bg-blue-50 border border-blue-200 rounded-xl p-3">
            <p className="text-xs text-blue-700">
              <span className="font-bold">必填字段：</span>
              {columns.filter(c => c.required).map(c => c.label).join('、')}
            </p>
          </div>

          {/* 操作按钮：下载模板、导出数据 */}
          <div className="flex gap-3">
            <button
              onClick={generateTemplate}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-slate-100 hover:bg-slate-200 rounded-xl text-xs font-black text-slate-600 transition-all"
            >
              <Download size={14} />
              下载导入模板
            </button>
            {onExport && (
              <button
                onClick={handleExport}
                disabled={status === 'importing'}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-emerald-100 hover:bg-emerald-200 rounded-xl text-xs font-black text-emerald-600 transition-all disabled:opacity-50"
              >
                <Download size={14} />
                导出当前数据
              </button>
            )}
          </div>

          {/* 文件选择 */}
          {status === 'idle' && (
            <button
              type="button"
              onClick={handleSelectFile}
              className="block w-full border-2 border-dashed border-slate-200 rounded-2xl p-8 text-center hover:border-blue-300 hover:bg-blue-50/30 transition-all cursor-pointer"
            >
              <FileSpreadsheet size={32} className="mx-auto mb-3 text-slate-300" />
              <p className="text-sm font-bold text-slate-600 mb-1">点击选择 Excel 文件</p>
              <p className="text-xs text-slate-400">支持 .xlsx 或 .xls 格式，最大 {maxSize}MB</p>
            </button>
          )}

          {/* 验证中 */}
          {status === 'validating' && (
            <div className="flex flex-col items-center py-8">
              <Loader2 size={32} className="animate-spin text-blue-500 mb-3" />
              <p className="text-sm font-bold text-slate-600">正在验证文件...</p>
            </div>
          )}

          {/* 验证错误 */}
          {status === 'error' && validationErrors.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-2xl p-4">
              <div className="flex items-center gap-2 mb-3">
                <XCircle size={18} className="text-red-500" />
                <span className="text-sm font-black text-red-700">文件验证失败</span>
              </div>
              <ul className="space-y-1 max-h-32 overflow-y-auto">
                {validationErrors.slice(0, 10).map((err, i) => (
                  <li key={i} className="text-xs text-red-600">
                    {err.row > 0 && `第 ${err.row} 行：`}
                    <span className="font-bold">{err.field}</span> - {err.message}
                  </li>
                ))}
                {validationErrors.length > 10 && (
                  <li className="text-xs text-red-400">... 还有 {validationErrors.length - 10} 条错误</li>
                )}
              </ul>
            </div>
          )}

          {/* 预览准备导入 */}
          {status === 'ready' && previewData && (
            <div className="bg-green-50 border border-green-200 rounded-2xl p-4">
              <div className="flex items-center gap-2 mb-3">
                <CheckCircle2 size={18} className="text-green-500" />
                <span className="text-sm font-black text-green-700">文件验证通过</span>
              </div>
              <p className="text-xs text-green-600 mb-2">
                文件：<span className="font-bold">{selectedFile?.name}</span>
              </p>
              <p className="text-xs text-green-600">
                共 <span className="font-bold">{totalDataCount}</span> 条数据
              </p>
            </div>
          )}

          {/* 确认添加缺失的药品分类 */}
          {status === 'confirm_categories' && (
            <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4">
              <div className="flex items-center gap-2 mb-3">
                <AlertCircle size={18} className="text-amber-500" />
                <span className="text-sm font-black text-amber-700">发现新的药品分类</span>
              </div>
              <p className="text-xs text-amber-600 mb-3">
                Excel 文件中包含以下药品分类在系统中不存在：
              </p>
              <ul className="space-y-1 mb-4 max-h-32 overflow-y-auto">
                {missingCategories.map((cat, i) => (
                  <li key={i} className="text-xs text-amber-700 bg-amber-100/50 rounded px-2 py-1">
                    {cat}
                  </li>
                ))}
              </ul>
              <p className="text-xs text-amber-600">
                是否将这些分类自动添加到系统字典中？
              </p>
            </div>
          )}

          {/* 导入中 */}
          {status === 'importing' && (
            <div className="flex flex-col items-center py-8">
              <Loader2 size={32} className="animate-spin text-blue-500 mb-3" />
              <p className="text-sm font-bold text-slate-600 mb-2">正在导入数据...</p>
              <div className="w-full bg-slate-100 rounded-full h-2 max-w-xs">
                <div
                  className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}

          {/* 导入成功 */}
          {status === 'success' && importResult && (
            <div className="bg-green-50 border border-green-200 rounded-2xl p-4">
              <div className="flex items-center gap-2 mb-3">
                <CheckCircle2 size={18} className="text-green-500" />
                <span className="text-sm font-black text-green-700">导入成功！</span>
              </div>
              <p className="text-xs text-green-600">
                成功导入 <span className="font-bold">{importResult.success}</span> 条数据
              </p>
              {(importResult.skipped ?? 0) > 0 && (
                <p className="text-xs text-amber-600 mt-1">
                  跳过 <span className="font-bold">{importResult.skipped}</span> 条重复数据
                </p>
              )}
            </div>
          )}

          {/* 部分成功 */}
          {(status === 'partial' || status === 'error') && importResult && (
            <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4">
              <div className="flex items-center gap-2 mb-3">
                <AlertCircle size={18} className="text-amber-500" />
                <span className="text-sm font-black text-amber-700">导入完成（部分失败）</span>
              </div>
              <p className="text-xs text-amber-600 mb-2">
                成功：<span className="font-bold">{importResult.success}</span> 条，
                失败：<span className="font-bold">{importResult.failed}</span> 条，
                跳过：<span className="font-bold">{importResult.skipped ?? 0}</span> 条
              </p>
              {importResult.errors.length > 0 && (
                <details className="mt-2">
                  <summary className="text-xs text-amber-600 cursor-pointer hover:underline">
                    查看错误详情 ({importResult.errors.length} 条)
                  </summary>
                  <ul className="mt-2 space-y-1 max-h-24 overflow-y-auto">
                    {importResult.errors.slice(0, 10).map((err, i) => (
                      <li key={i} className="text-xs text-amber-600">{err}</li>
                    ))}
                    {importResult.errors.length > 10 && (
                      <li className="text-xs text-amber-400">... 还有 {importResult.errors.length - 10} 条错误</li>
                    )}
                  </ul>
                </details>
              )}
            </div>
          )}
        </div>

        {/* 操作按钮 */}
        <div className="flex justify-end gap-3 mt-8">
          {status === 'error' && (
            <button
              onClick={handleReset}
              className="px-6 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-200 transition-all"
            >
              重新选择
            </button>
          )}

          {status === 'ready' && (
            <>
              <button
                onClick={handleReset}
                className="px-6 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-200 transition-all"
              >
                重新选择
              </button>
              <button
                onClick={handleImport}
                className="px-6 py-2 bg-blue-500 text-white rounded-xl text-xs font-black hover:bg-blue-600 transition-all flex items-center gap-2"
              >
                <Upload size={14} />
                开始导入
              </button>
            </>
          )}

          {status === 'confirm_categories' && (
            <>
              <button
                onClick={handleReset}
                className="px-6 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-200 transition-all"
              >
                取消
              </button>
              <button
                onClick={handleConfirmAddCategories}
                className="px-6 py-2 bg-emerald-500 text-white rounded-xl text-xs font-black hover:bg-emerald-600 transition-all flex items-center gap-2"
              >
                <CheckCircle2 size={14} />
                确认添加并导入
              </button>
            </>
          )}

          {(status === 'success' || status === 'partial') && (
            <button
              onClick={handleClose}
              className="px-6 py-2 bg-blue-500 text-white rounded-xl text-xs font-black hover:bg-blue-600 transition-all"
            >
              完成
            </button>
          )}

          {status === 'idle' && (
            <button
              onClick={handleClose}
              className="px-6 py-2 bg-slate-100 text-slate-600 rounded-xl text-xs font-black hover:bg-slate-200 transition-all"
            >
              取消
            </button>
          )}
        </div>
      </div>
    </div>
  )
})
