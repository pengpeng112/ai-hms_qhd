// PaginationBar.tsx - 分页 UI 组件

import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'

interface PaginationBarProps {
  currentPage: number
  totalPages: number
  displayInfo: {
    startIndex: number
    endIndex: number
    total: number
  }
  goToPage: (page: number) => void
  firstPage: () => void
  lastPage: () => void
  prevPage: () => void
  nextPage: () => void
}

export function PaginationBar({
  currentPage,
  totalPages,
  displayInfo,
  goToPage,
  firstPage,
  lastPage,
  prevPage,
  nextPage
}: PaginationBarProps) {
  // 生成页码按钮
  const getPageNumbers = () => {
    const pages: (number | string)[] = []
    const maxVisible = 5

    if (totalPages <= maxVisible) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i)
      }
    } else {
      if (currentPage <= 3) {
        pages.push(1, 2, 3, 4, '...', totalPages)
      } else if (currentPage >= totalPages - 2) {
        pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages)
      } else {
        pages.push(1, '...', currentPage - 1, currentPage, currentPage + 1, '...', totalPages)
      }
    }
    return pages
  }

  return (
    <div className="flex items-center justify-between px-6 py-4 bg-white border-t border-slate-100">
      <div className="text-xs font-medium text-slate-500">
        显示 {displayInfo.startIndex} - {displayInfo.endIndex} 条，共 {displayInfo.total} 条
      </div>
      <div className="flex items-center gap-2">
        <button
          onClick={firstPage}
          disabled={currentPage === 1}
          className="p-2 rounded-lg hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="首页"
        >
          <ChevronsLeft size={16} />
        </button>
        <button
          onClick={prevPage}
          disabled={currentPage === 1}
          className="p-2 rounded-lg hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="上一页"
        >
          <ChevronLeft size={16} />
        </button>
        <div className="flex gap-1">
          {getPageNumbers().map((page, index) => (
            typeof page === 'number' ? (
              <button
                key={index}
                onClick={() => goToPage(page)}
                className={`min-w-[32px] h-8 px-2 rounded-lg text-xs font-bold transition-all ${
                  currentPage === page
                    ? 'bg-blue-600 text-white'
                    : 'hover:bg-slate-100 text-slate-600'
                }`}
              >
                {page}
              </button>
            ) : (
              <span key={index} className="px-1 text-slate-400">
                {page}
              </span>
            )
          ))}
        </div>
        <button
          onClick={nextPage}
          disabled={currentPage === totalPages}
          className="p-2 rounded-lg hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="下一页"
        >
          <ChevronRight size={16} />
        </button>
        <button
          onClick={lastPage}
          disabled={currentPage === totalPages}
          className="p-2 rounded-lg hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="末页"
        >
          <ChevronsRight size={16} />
        </button>
      </div>
    </div>
  )
}
