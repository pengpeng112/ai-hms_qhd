const toNumberOrNull = (value?: string | number) => {
  if (value === undefined || value === null) return null
  const raw = typeof value === 'number' ? value : parseFloat(value)
  return Number.isFinite(raw) ? raw : null
}

const roundDown = (value: number, decimals = 2) => {
  const factor = 10 ** decimals
  return Math.floor(value * factor) / factor
}

const roundTo = (value: number, decimals = 2) => {
  const factor = 10 ** decimals
  return Math.round(value * factor) / factor
}

const formatNumber = (value: number, decimals = 2) => {
  const rounded = roundDown(value, decimals)
  return Number.isFinite(rounded) ? String(rounded) : ''
}

const formatNumberRounded = (value: number, decimals = 2) => {
  const rounded = roundTo(value, decimals)
  return Number.isFinite(rounded) ? String(rounded) : ''
}

export const calcSubstituteVolume = (durationHours?: string | number, flowMlMin?: string | number) => {
  const hours = toNumberOrNull(durationHours)
  const flow = toNumberOrNull(flowMlMin)
  if (hours === null || flow === null) return ''
  const volume = hours * flow * 60 / 1000
  return formatNumber(volume, 2)
}

export const calcDialysateVolume = (durationHours?: string | number, flowMlMin?: string | number) => {
  const hours = toNumberOrNull(durationHours)
  const flow = toNumberOrNull(flowMlMin)
  if (hours === null || flow === null) return ''
  const volume = hours * flow * 60 / 1000
  return formatNumber(volume, 2)
}

export const calcInjectionVolume = (rate?: string | number, duration?: string | number) => {
  const safeRate = toNumberOrNull(rate)
  const safeDuration = toNumberOrNull(duration)
  if (safeRate === null || safeDuration === null) return ''
  return formatNumber(safeRate * safeDuration, 2)
}

export const calcTotalDose = (maintenanceDose?: string | number, firstDose?: string | number) => {
  const maintenance = toNumberOrNull(maintenanceDose)
  const first = toNumberOrNull(firstDose)
  if (maintenance === null && first === null) return ''
  const total = (maintenance ?? 0) + (first ?? 0)
  if (total === 0) return ''
  return formatNumberRounded(total, 2)
}
