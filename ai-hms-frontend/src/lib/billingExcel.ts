import * as XLSX from 'xlsx'
import type { ChargeRecord } from '@services/billingApi'

const CATEGORY_LABELS: Record<string, string> = {
  treatment: 'A 治疗费',
  material: 'B 耗材费',
  nursing: 'C 护理费',
  injection: 'D 注射费',
  drug: 'E 药品（HIS查价）',
}

export function exportChargeToExcel(record: ChargeRecord) {
  const lines = record.lines ?? []
  const header = ['类别', '项目名称', '规格', '单位', '数量', '参考单价', '参考金额', '医保编码', 'HIS编码', '匹配状态', '备注']

  const rows = lines.map((l) => {
    const showPrice = l.unitPrice != null && l.billable && l.category !== 'drug'
    return [
      CATEGORY_LABELS[l.category] ?? l.category,
      l.itemName,
      l.spec ?? '',
      l.unit ?? '',
      l.quantity ?? 1,
      showPrice ? l.unitPrice : (l.category === 'drug' ? '' : '-'),
      showPrice ? l.amount : (l.category === 'drug' ? '' : '-'),
      l.itemCode ?? '',
      l.hisItemCode ?? '',
      l.matchedStatus ?? '',
      l.note ?? '',
    ]
  })

  rows.push([
    '', '参考总价（仅供核对，实际以HIS为准）', '', '', '', '',
    record.totalAmount ?? 0, '', '', '', '',
  ])

  const ws = XLSX.utils.aoa_to_sheet([header, ...rows])
  ws['!cols'] = [
    { wch: 12 }, { wch: 28 }, { wch: 12 }, { wch: 6 },
    { wch: 8 }, { wch: 10 }, { wch: 10 }, { wch: 18 }, { wch: 18 }, { wch: 10 }, { wch: 20 },
  ]

  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, '收费清单')
  XLSX.writeFile(wb, `收费清单_${record.chargeDate ?? record.id}.xlsx`)
}
