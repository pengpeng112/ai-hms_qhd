import { useState, useEffect, useCallback } from 'react'
import { restApi, type SyncJobConfig, type SyncJobRun, type UnmatchedPatientItem, type UnmatchedPatientResponse, type PatientSearchItem } from '@/services/restClient'
import { getErrorMessage } from '@/services/restClient'
import { Play, RefreshCw, Link2, Search, Check, ChevronLeft, ChevronRight } from 'lucide-react'

export default function SyncCenterPage() {
  const [jobs, setJobs] = useState<SyncJobConfig[]>([])
  const [runs, setRuns] = useState<Record<string, SyncJobRun[]>>({})
  const [running, setRunning] = useState<Record<string, boolean>>({})
  const [loading, setLoading] = useState(true)

  const [unmatchedItems, setUnmatchedItems] = useState<UnmatchedPatientItem[]>([])
  const [unmatchedTotal, setUnmatchedTotal] = useState(0)
  const [unmatchedPage, setUnmatchedPage] = useState(1)
  const [unmatchedPageSize] = useState(20)
  const [unmatchedKeyword, setUnmatchedKeyword] = useState('')
  const [unmatchedLoading, setUnmatchedLoading] = useState(false)

  const [bindModal, setBindModal] = useState<{ open: boolean; externalId: string }>({ open: false, externalId: '' })
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchResults, setSearchResults] = useState<PatientSearchItem[]>([])
  const [searching, setSearching] = useState(false)
  const [binding, setBinding] = useState(false)

  const loadJobs = useCallback(async () => {
    try {
      const list = await restApi.getSyncJobs()
      setJobs(list)
    } catch { /**/ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { void loadJobs() }, [loadJobs])

  const handleRun = async (code: string) => {
    setRunning(prev => ({ ...prev, [code]: true }))
    try {
      await restApi.runSyncJob(code)
      setTimeout(() => { handleLoadRuns(code); loadJobs() }, 3000)
    } catch (e) { alert(getErrorMessage(e)) }
    finally { setRunning(prev => ({ ...prev, [code]: false })) }
  }

  const handleLoadRuns = async (code: string) => {
    try {
      const list = await restApi.getSyncJobRuns(code)
      setRuns(prev => ({ ...prev, [code]: list }))
    } catch { /**/ }
  }

  const handleToggle = async (code: string, enabled: boolean) => {
    try {
      await restApi.updateSyncJob(code, { enabled: !enabled })
      loadJobs()
    } catch (e) { alert(getErrorMessage(e)) }
  }

  const loadUnmatched = async (page?: number, keyword?: string) => {
    setUnmatchedLoading(true)
    const p = page ?? unmatchedPage
    const kw = keyword !== undefined ? keyword : unmatchedKeyword
    try {
      const res: UnmatchedPatientResponse = await restApi.getUnmatchedPatients({
        page: p,
        pageSize: unmatchedPageSize,
        keyword: kw || undefined,
      })
      setUnmatchedItems(res.items)
      setUnmatchedTotal(res.total)
      setUnmatchedPage(p)
    } catch (e) { alert(getErrorMessage(e)) }
    finally { setUnmatchedLoading(false) }
  }

  const handleUnmatchedSearch = () => {
    setUnmatchedPage(1)
    loadUnmatched(1, unmatchedKeyword)
  }

  const handleUnmatchedPageChange = (page: number) => {
    loadUnmatched(page, unmatchedKeyword)
  }

  const openBindModal = (externalId: string) => {
    setBindModal({ open: true, externalId })
    setSearchKeyword('')
    setSearchResults([])
  }

  const handleSearch = async () => {
    if (!searchKeyword.trim()) return
    setSearching(true)
    try {
      const results = await restApi.searchPatients(searchKeyword)
      setSearchResults(results)
    } catch { /**/ }
    finally { setSearching(false) }
  }

  const handleBind = async (legacyPatientId: number) => {
    setBinding(true)
    try {
      await restApi.bindExternalPatientMapping(bindModal.externalId, legacyPatientId)
      setBindModal({ open: false, externalId: '' })
      loadUnmatched(unmatchedPage, unmatchedKeyword)
    } catch (e) { alert(getErrorMessage(e)) }
    finally { setBinding(false) }
  }

  const statusBadge = (status: string) => {
    const map: Record<string, string> = { success: 'bg-green-100 text-green-600', partial: 'bg-amber-100 text-amber-600', failed: 'bg-red-100 text-red-600', running: 'bg-blue-100 text-blue-600' }
    return `px-2 py-0.5 rounded text-[10px] font-medium ${map[status] || 'bg-gray-100 text-gray-500'}`
  }

  const totalPages = Math.max(1, Math.ceil(unmatchedTotal / unmatchedPageSize))

  if (loading) return <div className="p-8 text-gray-400">加载中...</div>

  return (
    <div className="p-6 max-w-6xl">
      <h2 className="text-xl font-bold text-gray-800 mb-6">同步管理中心</h2>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="border border-gray-200 rounded-lg p-5">
          <h3 className="font-semibold text-gray-800 mb-4 flex items-center gap-2"><RefreshCw size={18} /> 同步任务</h3>
          <div className="space-y-3">
            {jobs.map((job) => (
              <div key={job.jobCode} className="p-4 border border-gray-100 rounded-lg bg-gray-50/50">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-semibold text-gray-700">{job.jobCode}</span>
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${job.enabled ? 'bg-green-100 text-green-600' : 'bg-gray-200 text-gray-500'}`}>
                    {job.enabled ? '启用' : '停用'}
                  </span>
                </div>
                <div className="text-xs text-gray-500 space-y-1">
                  <div>每批 {job.batchSize} 条 · 超时 {job.timeoutSeconds}s</div>
                  {job.lastRunAt && <div>上次运行: {new Date(job.lastRunAt).toLocaleString()}</div>}
                </div>
                <div className="flex gap-2 mt-3">
                  <button
                    onClick={() => handleToggle(job.jobCode, job.enabled)}
                    className={`px-3 py-1 rounded text-xs font-medium ${job.enabled ? 'bg-gray-100 text-gray-600 hover:bg-gray-200' : 'bg-green-50 text-green-600 hover:bg-green-100'}`}
                  >
                    {job.enabled ? '停用' : '启用'}
                  </button>
                  <button
                    disabled={!job.enabled || running[job.jobCode]}
                    onClick={() => handleRun(job.jobCode)}
                    className={`px-3 py-1 rounded text-xs font-medium inline-flex items-center gap-1 ${!job.enabled ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-blue-50 text-blue-600 hover:bg-blue-100'}`}
                  >
                    <Play size={12} className={running[job.jobCode] ? 'animate-spin' : ''} />
                    立即运行
                  </button>
                  <button
                    onClick={() => handleLoadRuns(job.jobCode)}
                    className="px-3 py-1 rounded text-xs font-medium bg-white border border-gray-200 text-gray-600 hover:bg-gray-50"
                  >
                    历史
                  </button>
                </div>
                {runs[job.jobCode] && runs[job.jobCode].length > 0 && (
                  <div className="mt-3 border-t border-gray-100 pt-2">
                    <table className="w-full text-[10px]">
                      <thead>
                        <tr className="text-gray-400">
                          <th className="text-left pb-1">时间</th>
                          <th className="text-left pb-1">状态</th>
                          <th className="text-right pb-1">获取</th>
                          <th className="text-right pb-1">新增</th>
                          <th className="text-right pb-1">更新</th>
                          <th className="text-right pb-1">跳过</th>
                          <th className="text-right pb-1">失败</th>
                        </tr>
                      </thead>
                      <tbody>
                        {runs[job.jobCode].slice(0, 3).map((r) => (
                          <tr key={r.id}>
                            <td className="py-0.5 text-gray-500">{new Date(r.startedAt).toLocaleTimeString()}</td>
                            <td className="py-0.5"><span className={statusBadge(r.status)}>{r.status}</span></td>
                            <td className="py-0.5 text-right text-gray-600">{r.fetchedCount}</td>
                            <td className="py-0.5 text-right text-green-600">{r.createdCount}</td>
                            <td className="py-0.5 text-right text-blue-600">{r.updatedCount}</td>
                            <td className="py-0.5 text-right text-gray-400">{r.skippedCount}</td>
                            <td className="py-0.5 text-right text-red-500">{r.failedCount}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        <div className="border border-gray-200 rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold text-gray-800 flex items-center gap-2"><Link2 size={18} /> 未匹配患者</h3>
            <button
              onClick={() => { setUnmatchedPage(1); setUnmatchedKeyword(''); loadUnmatched(1, '') }}
              disabled={unmatchedLoading}
              className="px-3 py-1 rounded text-xs font-medium bg-white border border-gray-200 text-gray-600 hover:bg-gray-50"
            >
              {unmatchedLoading ? '加载中...' : '刷新'}
            </button>
          </div>

          <div className="flex gap-2 mb-3">
            <input
              type="text"
              value={unmatchedKeyword}
              onChange={(e) => setUnmatchedKeyword(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleUnmatchedSearch() }}
              placeholder="搜索 HIS 患者号 / 姓名"
              className="flex-1 px-3 py-1.5 border border-gray-300 rounded text-xs outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleUnmatchedSearch}
              disabled={unmatchedLoading}
              className="px-3 py-1.5 bg-blue-600 text-white rounded text-xs font-medium hover:bg-blue-700 disabled:opacity-60"
            >
              搜索
            </button>
          </div>

          {unmatchedItems.length === 0 ? (
            <div className="text-sm text-gray-400 py-6 text-center">
              {unmatchedLoading ? '查询 HIS 中...' : '点击「刷新」或搜索查询 HIS 中未匹配的患者'}
            </div>
          ) : (
            <>
              <div className="space-y-2 max-h-[380px] overflow-y-auto">
                {unmatchedItems.map((p) => (
                  <div key={p.patientId} className="flex items-center justify-between p-3 border border-gray-100 rounded-lg bg-gray-50/50">
                    <div>
                      <div className="text-sm font-medium text-gray-700">{p.name}</div>
                      <div className="text-[10px] text-gray-400">HIS: {p.patientId} · 检查数: {p.examCnt}</div>
                    </div>
                    <button
                      onClick={() => openBindModal(p.patientId)}
                      className="px-3 py-1 rounded text-xs font-medium bg-blue-50 text-blue-600 hover:bg-blue-100 inline-flex items-center gap-1"
                    >
                      <Search size={12} />
                      绑定
                    </button>
                  </div>
                ))}
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-100">
                  <span className="text-[10px] text-gray-400">共 {unmatchedTotal} 个患者 · 第 {unmatchedPage}/{totalPages} 页</span>
                  <div className="flex gap-1">
                    <button
                      disabled={unmatchedPage <= 1}
                      onClick={() => handleUnmatchedPageChange(unmatchedPage - 1)}
                      className="px-2 py-1 rounded text-xs bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <ChevronLeft size={14} />
                    </button>
                    <button
                      disabled={unmatchedPage >= totalPages}
                      onClick={() => handleUnmatchedPageChange(unmatchedPage + 1)}
                      className="px-2 py-1 rounded text-xs bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      <ChevronRight size={14} />
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {bindModal.open && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setBindModal({ open: false, externalId: '' })}>
          <div className="bg-white rounded-lg p-6 w-full max-w-md shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="font-semibold text-gray-800 mb-4">绑定本地患者</h3>
            <p className="text-xs text-gray-500 mb-3">HIS 患者: <span className="font-mono text-gray-700">{bindModal.externalId}</span></p>
            <div className="flex gap-2 mb-4">
              <input
                type="text"
                value={searchKeyword}
                onChange={(e) => setSearchKeyword(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleSearch() }}
                placeholder="搜索姓名 / 透析号 / ID"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={handleSearch}
                disabled={searching}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-60"
              >
                {searching ? '搜索中...' : '搜索'}
              </button>
            </div>
            {searchResults.length > 0 && (
              <div className="space-y-2 max-h-[300px] overflow-y-auto">
                {searchResults.map((p) => (
                  <div key={p.id} className="flex items-center justify-between p-3 border border-gray-100 rounded-lg">
                    <div>
                      <div className="text-sm font-medium text-gray-700">{p.name}</div>
                      <div className="text-[10px] text-gray-400">{p.gender} · {p.age}岁 · ID: {p.id}</div>
                      {p.dialysisNo && <div className="text-[10px] text-gray-400">透析号: {p.dialysisNo}</div>}
                    </div>
                    <button
                      onClick={() => handleBind(p.id)}
                      disabled={binding}
                      className="px-3 py-1 rounded text-xs font-medium bg-green-50 text-green-600 hover:bg-green-100 inline-flex items-center gap-1"
                    >
                      <Check size={12} />
                      绑定
                    </button>
                  </div>
                ))}
              </div>
            )}
            {searchResults.length === 0 && searchKeyword && !searching && (
              <div className="text-sm text-gray-400 py-4 text-center">未找到匹配患者</div>
            )}
            <button
              onClick={() => setBindModal({ open: false, externalId: '' })}
              className="w-full mt-4 px-4 py-2 bg-gray-100 text-gray-600 rounded-lg text-sm hover:bg-gray-200"
            >
              取消
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
