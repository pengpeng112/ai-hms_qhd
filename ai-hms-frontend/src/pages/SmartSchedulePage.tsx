import { useState, useEffect, useCallback } from 'react'
import { DatePicker, Button, Card, Modal, message, Spin, Tabs, Table, Tag, Statistic, Row, Col } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import {
  getBoard,
  generateSchedule,
  confirmPlan,
  confirmDay,
  cancelShift,
  listConflicts,
  resolveConflict,
  getQuality,
  getDiffs,
  listIncompleteProfiles,
  type WeekBoard,
  type CellDTO,
  type MachineDTO,
  type WardDTO,
  type ConflictItem,
  type QualityResult,
  type DiffItem,
  type IncompleteItem,
} from '@/services/smartScheduleApi'

export default function SmartSchedulePage() {
  const [currentDate, setCurrentDate] = useState<Dayjs>(dayjs())
  const [board, setBoard] = useState<WeekBoard | null>(null)
  const [loading, setLoading] = useState(false)
  const [conflicts, setConflicts] = useState<ConflictItem[]>([])
  const [quality, setQuality] = useState<QualityResult | null>(null)
  const [diffs, setDiffs] = useState<DiffItem[]>([])
  const [incomplete, setIncomplete] = useState<IncompleteItem[]>([])

  const fetchData = useCallback(async () => {
    const dateStr = currentDate.format('YYYY-MM-DD')
    setLoading(true)
    try {
      const [boardRes, conflictsRes, qualityRes, diffsRes, incompleteRes] = await Promise.allSettled([
        getBoard(dateStr),
        listConflicts(0),
        getQuality(dateStr, 2),
        getDiffs(dateStr, 2),
        listIncompleteProfiles(),
      ])
      if (boardRes.status === 'fulfilled') setBoard(boardRes.value)
      if (conflictsRes.status === 'fulfilled') setConflicts(conflictsRes.value.conflicts ?? [])
      if (qualityRes.status === 'fulfilled') setQuality(qualityRes.value)
      if (diffsRes.status === 'fulfilled') setDiffs(diffsRes.value.items ?? [])
      if (incompleteRes.status === 'fulfilled') setIncomplete(incompleteRes.value.items ?? [])
    } finally {
      setLoading(false)
    }
  }, [currentDate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleGenerate = async () => {
    try {
      const res = await generateSchedule({ startDate: currentDate.format('YYYY-MM-DD'), weeks: 2 })
      message.success(`生成完成: ${res.drafts} 条草稿, ${res.conflicts} 条冲突`)
      fetchData()
    } catch {
      message.error('生成失败')
    }
  }

  const handleConfirmPlan = async () => {
    try {
      const res = await confirmPlan({ weekStart: board?.weekStart, weeks: 2 })
      message.success(`已确认 ${res.confirmed} 条排班`)
      fetchData()
    } catch {
      message.error('确认失败')
    }
  }

  const handleConfirmDay = async (level: number) => {
    try {
      const res = await confirmDay({ date: currentDate.format('YYYY-MM-DD'), level })
      message.success(`${level === 2 ? '次日' : '当日'}确认: ${res.confirmed} 条`)
      fetchData()
    } catch {
      message.error('确认失败')
    }
  }

  const handleCancelShift = async (id: number) => {
    try {
      await cancelShift(id, '人工取消')
      message.success('已取消')
      fetchData()
    } catch {
      message.error('取消失败')
    }
  }

  const cellClick = (cell: CellDTO, _machine: MachineDTO, _date: string) => {
    if (!cell || cell.id === 0) {
      return // 空机位无操作，临时排班通过顶部按钮
    }
    Modal.confirm({
      title: `${cell.patientName} - ${cell.dialysisMode}`,
      content: `状态: ${cell.status}, 确认级数: ${cell.confirms}`,
      okText: '取消此排班',
      onOk: () => handleCancelShift(cell.id),
    })
  }

  const getShiftCells = (machine: MachineDTO, shiftId: number): (CellDTO | null)[] => {
    if (!board || !board.dates) return []
    return board.dates.map((date) => {
      const key = `${date}|${shiftId}`
      return machine.cells?.[key] ?? { id: 0, shiftId: 0, patientId: 0, patientName: '', dialysisMode: '', status: 0, sourceType: 0, confirms: 0 }
    })
  }

  return (
    <div className="p-4">
      <h1 className="text-xl font-bold mb-4">智能排班</h1>

      <div className="flex gap-3 mb-4 flex-wrap items-center">
        <DatePicker value={currentDate} onChange={(d) => d && setCurrentDate(d)} allowClear={false} />
        <Button type="primary" onClick={handleGenerate}>生成排班</Button>
        <Button onClick={handleConfirmPlan}>整盘确认</Button>
        <Button onClick={() => handleConfirmDay(2)}>次日确认</Button>
        <Button onClick={() => handleConfirmDay(3)}>当日确认</Button>
        <Button onClick={fetchData}>刷新</Button>
      </div>

      <Spin spinning={loading}>
        {quality?.onTargetRate !== undefined && (
          <Row gutter={16} className="mb-4">
            <Col span={4}><Card size="small"><Statistic title="达标率" value={quality.onTargetRate} suffix="%" precision={0} valueStyle={{ color: quality.onTargetRate >= 80 ? '#3f8600' : '#cf1322' }} /></Card></Col>
            <Col span={4}><Card size="small"><Statistic title="利用率" value={quality.utilization} suffix="%" precision={0} /></Card></Col>
            <Col span={4}><Card size="small"><Statistic title="稳定率" value={quality.stabilityRate} suffix="%" precision={0} /></Card></Col>
            <Col span={4}><Card size="small"><Statistic title="综合分" value={quality.score} suffix="/100" /></Card></Col>
            <Col span={4}><Card size="small"><Statistic title="待处理冲突" value={quality.openConflicts} valueStyle={{ color: quality.openConflicts > 0 ? '#cf1322' : '#3f8600' }} /></Card></Col>
            <Col span={4}><Card size="small"><Statistic title="达标患者" value={`${quality.patientsOnTarget}/${quality.patientsTotal}`} /></Card></Col>
          </Row>
        )}

        <Tabs
          items={[
            {
              key: 'board',
              label: '周排班矩阵',
              children: board ? (
                <div className="overflow-x-auto">
                  <table className="border-collapse w-full text-sm">
                    <thead>
                      <tr className="bg-gray-50">
                        <th className="border p-2">病区</th>
                        <th className="border p-2">机器</th>
                        {board.shifts?.map((s) => (
                          <th key={s.id} colSpan={board.dates?.length} className="border p-2 bg-blue-50">
                            {s.name}
                          </th>
                        ))}
                      </tr>
                      <tr className="bg-gray-50">
                        <th className="border p-2"></th>
                        <th className="border p-2"></th>
                        {board.shifts?.map((s) =>
                          board.dates?.map((d) => (
                            <th key={`${s.id}-${d}`} className="border p-1 text-xs font-normal">
                              {d.slice(5)}
                            </th>
                          ))
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {board.wards?.map((ward: WardDTO) =>
                        ward.machines?.map((machine: MachineDTO, mi: number) => (
                          <tr key={`${ward.id}-${machine.id}`}>
                            {mi === 0 && <td rowSpan={ward.machines.length} className="border p-2 font-bold align-top">{ward.name}</td>}
                            <td className="border p-1 text-xs">{machine.code}</td>
                            {board.shifts?.map((s) =>
                              getShiftCells(machine, s.id).map((cell, ci) => (
                                <td
                                  key={`${s.id}-${ci}`}
                                  className={`border p-1 text-center cursor-pointer text-xs min-w-[60px] ${
                                    cell && cell.status > 0 ? 'bg-green-50 hover:bg-green-100' : 'hover:bg-gray-100'
                                  }`}
                                  onClick={() => cellClick(cell ?? { id: 0, shiftId: s.id, patientId: 0, patientName: '', dialysisMode: '', status: 0, sourceType: 0, confirms: 0 }, machine, board.dates[ci])}
                                >
                                  {cell && cell.status > 0 ? (
                                    <div>
                                      <div className="font-bold">{cell.patientName}</div>
                                      <div className="text-xs text-gray-500">{cell.dialysisMode}</div>
                                      {cell.confirms > 0 && <Tag color="blue" className="text-[10px] leading-none px-1">{'✓'.repeat(cell.confirms)}</Tag>}
                                    </div>
                                  ) : (
                                    <div className="text-gray-300">空</div>
                                  )}
                                </td>
                              ))
                            )}
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              ) : <div className="text-gray-400 text-center py-8">暂无排班数据</div>,
            },
            {
              key: 'conflicts',
              label: `冲突队列 (${conflicts.length})`,
              children: conflicts.length > 0 ? (
                <Table
                  dataSource={conflicts}
                  rowKey="id"
                  size="small"
                  columns={[
                    { title: '类型', dataIndex: 'conflictType', key: 'type', render: (v: string) => <Tag>{v}</Tag> },
                    { title: '严重度', dataIndex: 'severity', key: 'severity', render: (v: number) => <Tag color={v >= 20 ? 'red' : 'orange'}>{v >= 20 ? '报警' : '提示'}</Tag> },
                    { title: '详情', dataIndex: 'detail', key: 'detail' },
                    { title: '状态', dataIndex: 'status', key: 'status', render: (v: number) => v === 0 ? '待处理' : v === 10 ? '已处理' : '已忽略' },
                    { title: '操作', key: 'action', render: (_: unknown, r: ConflictItem) => (
                      <span>
                        <Button size="small" type="link" onClick={async () => { await resolveConflict(r.id, 'accept'); message.success('已处理'); fetchData() }}>接受</Button>
                        <Button size="small" type="link" onClick={async () => { await resolveConflict(r.id, 'ignore'); message.success('已忽略'); fetchData() }}>忽略</Button>
                      </span>
                    )},
                  ]}
                />
              ) : <div className="text-gray-400 text-center py-4">无待处理冲突</div>,
            },
            {
              key: 'diffs',
              label: `应排差异 (${diffs.length})`,
              children: diffs.length > 0 ? (
                <Table
                  dataSource={diffs}
                  rowKey="patientId"
                  size="small"
                  columns={[
                    { title: '病人', dataIndex: 'patientName', key: 'name' },
                    { title: '应排', dataIndex: 'expected', key: 'expected' },
                    { title: '已排', dataIndex: 'scheduled', key: 'scheduled' },
                    { title: '差异', dataIndex: 'diff', key: 'diff', render: (v: number) => <Tag color={v > 0 ? 'red' : 'green'}>{v > 0 ? `少 ${v}` : `多 ${-v}`}</Tag> },
                  ]}
                />
              ) : <div className="text-gray-400 text-center py-4">无应排差异</div>,
            },
            {
              key: 'incomplete',
              label: `资料待补 (${incomplete.length})`,
              children: incomplete.length > 0 ? (
                <Table
                  dataSource={incomplete}
                  rowKey="patientId"
                  size="small"
                  columns={[
                    { title: '病人', dataIndex: 'patientName', key: 'name' },
                    { title: '缺少字段', dataIndex: 'missing', key: 'missing', render: (v: string[]) => v.map((m) => <Tag key={m} color="orange">{m}</Tag>) },
                  ]}
                />
              ) : <div className="text-gray-400 text-center py-4">所有病人资料完整</div>,
            },
          ]}
        />
      </Spin>
    </div>
  )
}
