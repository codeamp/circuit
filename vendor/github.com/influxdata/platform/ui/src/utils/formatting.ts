import _ from 'lodash'
const KMB_LABELS: string[] = ['K', 'M', 'B', 'T', 'Q']
const KMG2_BIG_LABELS: string[] = ['k', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y']
const KMG2_SMALL_LABELS: string[] = ['m', 'u', 'n', 'p', 'f', 'a', 'z', 'y']

export interface Duration {
  days: number
  hours: number
  minutes: number
  seconds: number
}

export const durationToSeconds = ({
  days,
  hours,
  minutes,
  seconds,
}: Duration): number => {
  const sInMinute = 60
  const sInHour = sInMinute * 60
  const sInDay = sInHour * 24

  const sDays = sInDay * days
  const sHours = sInHour * hours
  const sMinutes = sInMinute * minutes

  return sDays + sHours + sMinutes + seconds
}

export const secondsToDuration = (seconds: number): Duration => {
  let minutes = Math.floor(seconds / 60)
  seconds = seconds % 60
  let hours = Math.floor(minutes / 60)
  minutes = minutes % 60
  const days = Math.floor(hours / 24)
  hours = hours % 24

  return {
    days,
    hours,
    minutes,
    seconds,
  }
}

export const ruleToString = (seconds: number): string => {
  const duration = secondsToDuration(seconds)
  const rpString = Object.entries(duration).reduce((acc, [k, v]) => {
    if (!v) {
      return acc
    }

    return `${acc} ${v} ${k}`
  }, '')

  if (!rpString) {
    return 'forever'
  }

  return rpString
}

const pow = (base: number, exp: number): number => {
  if (exp < 0) {
    return 1.0 / Math.pow(base, -exp)
  }

  return Math.pow(base, exp)
}

const roundNum = (num, places): number => {
  const shift = Math.pow(10, places)
  return Math.round(num * shift) / shift
}

const floatFormat = (x: number, optPrecision: number): string => {
  // Avoid invalid precision values; [1, 21] is the valid range.
  const p = Math.min(Math.max(1, optPrecision || 2), 21)

  // This is deceptively simple.  The actual algorithm comes from:
  //
  // Max allowed length = p + 4
  // where 4 comes from 'e+n' and '.'.
  //
  // Length of fixed format = 2 + y + p
  // where 2 comes from '0.' and y = # of leading zeroes.
  //
  // Equating the two and solving for y yields y = 2, or 0.00xxxx which is
  // 1.0e-3.
  //
  // Since the behavior of toPrecision() is identical for larger numbers, we
  // don't have to worry about the other bound.
  //
  // Finally, the argument for toExponential() is the number of trailing digits,
  // so we take off 1 for the value before the '.'.
  return Math.abs(x) < 1.0e-3 && x !== 0.0
    ? x.toExponential(p - 1)
    : x.toPrecision(p)
}

// taken from https://github.com/danvk/dygraphs/blob/aaec6de56dba8ed712fd7b9d949de47b46a76ccd/src/dygraph-utils.js#L1103
export const numberValueFormatter = (
  x: number,
  opts: (name: string) => number,
  prefix: string,
  suffix: string
): string => {
  const sigFigs = opts('sigFigs')

  if (sigFigs !== null) {
    // User has opted for a fixed number of significant figures.
    return floatFormat(x, sigFigs)
  }

  const digits = opts('digitsAfterDecimal')
  const maxNumberWidth = opts('maxNumberWidth')

  const kmb = opts('labelsKMB')
  const kmg2 = opts('labelsKMG2')

  let label

  // switch to scientific notation if we underflow or overflow fixed display.
  if (
    x !== 0.0 &&
    (Math.abs(x) >= Math.pow(10, maxNumberWidth) ||
      Math.abs(x) < Math.pow(10, -digits))
  ) {
    label = x.toExponential(digits)
  } else {
    label = `${roundNum(x, digits)}`
  }

  if (kmb || kmg2) {
    let k
    let kLabels = []
    let mLabels = []
    if (kmb) {
      k = 1000
      kLabels = KMB_LABELS
    }
    if (kmg2) {
      if (kmb) {
        console.error('Setting both labelsKMB and labelsKMG2. Pick one!')
      }
      k = 1024
      kLabels = KMG2_BIG_LABELS
      mLabels = KMG2_SMALL_LABELS
    }

    const absx = Math.abs(x)
    let n = pow(k, kLabels.length)
    for (let j = kLabels.length - 1; j >= 0; j -= 1, n /= k) {
      if (absx >= n) {
        label = roundNum(x / n, digits) + kLabels[j]
        break
      }
    }
    if (kmg2) {
      const xParts = String(x.toExponential())
        .split('e-')
        .map(Number)
      if (xParts.length === 2 && xParts[1] >= 3 && xParts[1] <= 24) {
        if (xParts[1] % 3 > 0) {
          label = roundNum(xParts[0] / pow(10, xParts[1] % 3), digits)
        } else {
          label = Number(xParts[0]).toFixed(2)
        }
        label += mLabels[Math.floor(xParts[1] / 3) - 1]
      }
    }
  }

  return `${prefix}${label}${suffix}`
}
