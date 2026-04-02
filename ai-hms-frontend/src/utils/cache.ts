/**
 * API 缓存工具
 * 纯前端内存缓存，无外部依赖
 */

interface CacheItem<T> {
  data: T
  expireAt: number
  createdAt: number
}

// 缓存 TTL 配置（毫秒）
export const CACHE_TTL = {
  // 基础数据 - 变化少，缓存时间长
  WARD_LIST: 60 * 60 * 1000,        // 1 小时
  BED_LIST: 30 * 60 * 1000,         // 30 分钟
  SHIFT_LIST: 10 * 60 * 1000,       // 10 分钟
  EQUIPMENT_LIST: 30 * 60 * 1000,   // 30 分钟

  // 患者数据 - 适中缓存
  PATIENT_LIST: 5 * 60 * 1000,      // 5 分钟
  PATIENT_DETAIL: 5 * 60 * 1000,    // 5 分钟
  PATIENT_FULL_INFO: 5 * 60 * 1000, // 5 分钟

  // 动态数据 - 短缓存
  TODAY_SCHEDULE: 2 * 60 * 1000,    // 2 分钟
  TODAY_TREATMENTS: 1 * 60 * 1000,  // 1 分钟
  PATIENT_TREATMENTS: 2 * 60 * 1000, // 2 分钟

  // 默认
  DEFAULT: 5 * 60 * 1000,           // 5 分钟
} as const

class ApiCache {
  private cache = new Map<string, CacheItem<unknown>>()
  private maxSize = 500 // 最大缓存条目数

  /**
   * 生成缓存 key
   */
  static key(prefix: string, ...args: (string | number | undefined)[]): string {
    const parts = [prefix, ...args.filter(a => a !== undefined)]
    return parts.join(':')
  }

  /**
   * 获取缓存
   */
  get<T>(key: string): T | null {
    const item = this.cache.get(key)
    if (!item) return null

    // 检查是否过期
    if (Date.now() > item.expireAt) {
      this.cache.delete(key)
      return null
    }

    return item.data as T
  }

  /**
   * 设置缓存
   */
  set<T>(key: string, data: T, ttlMs: number = CACHE_TTL.DEFAULT): void {
    // 如果缓存已满，清理过期项
    if (this.cache.size >= this.maxSize) {
      this.cleanup()
    }

    // 如果清理后仍然满，删除最旧的项
    if (this.cache.size >= this.maxSize) {
      const oldestKey = this.cache.keys().next().value
      if (oldestKey) this.cache.delete(oldestKey)
    }

    this.cache.set(key, {
      data,
      expireAt: Date.now() + ttlMs,
      createdAt: Date.now()
    })
  }

  /**
   * 带缓存的异步函数包装器
   */
  async withCache<T>(
    key: string,
    fetcher: () => Promise<T>,
    ttlMs: number = CACHE_TTL.DEFAULT
  ): Promise<T> {
    // 检查缓存
    const cached = this.get<T>(key)
    if (cached !== null) {
      console.log(`[Cache HIT] ${key}`)
      return cached
    }

    // 获取数据
    console.log(`[Cache MISS] ${key}`)
    const data = await fetcher()

    // 存入缓存
    this.set(key, data, ttlMs)

    return data
  }

  /**
   * 删除匹配的缓存
   */
  invalidate(pattern?: string): void {
    if (!pattern) {
      this.cache.clear()
      console.log('[Cache] Cleared all')
      return
    }

    let count = 0
    for (const key of this.cache.keys()) {
      if (key.includes(pattern)) {
        this.cache.delete(key)
        count++
      }
    }
    console.log(`[Cache] Invalidated ${count} items matching "${pattern}"`)
  }

  /**
   * 删除指定 key 的缓存
   */
  delete(key: string): void {
    this.cache.delete(key)
  }

  /**
   * 清理过期缓存
   */
  cleanup(): void {
    const now = Date.now()
    let count = 0

    for (const [key, item] of this.cache.entries()) {
      if (now > item.expireAt) {
        this.cache.delete(key)
        count++
      }
    }

    if (count > 0) {
      console.log(`[Cache] Cleaned up ${count} expired items`)
    }
  }

  /**
   * 获取缓存统计
   */
  stats(): { size: number; keys: string[] } {
    return {
      size: this.cache.size,
      keys: Array.from(this.cache.keys())
    }
  }
}

// 单例导出
export const apiCache = new ApiCache()

// 便捷方法导出
export const cacheKey = ApiCache.key
