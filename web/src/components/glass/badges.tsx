import { gradeStep, hueColor, rampHue, type Flag } from '@/lib/grades'
import { cn } from '@/lib/utils'

/** A grade on the shared ramp. Grades with no place on it stay neutral. */
export function GradeBadge({ grade, size = 'md', className }: { grade: string; size?: 'sm' | 'md' | 'lg'; className?: string }) {
  const step = gradeStep(grade)
  const style =
    step === null
      ? undefined
      : { background: `oklch(0.86 0.12 ${rampHue(step)})`, color: `oklch(0.27 0.06 ${rampHue(step)})` }
  return (
    <span
      style={style}
      className={cn(
        'tabular inline-flex shrink-0 items-center justify-center rounded-md font-semibold',
        step === null && 'glass-sunk text-foreground',
        size === 'sm' && 'h-6 min-w-6 px-1.5 text-xs',
        size === 'md' && 'h-8 min-w-8 px-2 text-sm',
        size === 'lg' && 'h-12 min-w-12 rounded-lg px-3 text-xl',
        className,
      )}
    >
      <span className="sr-only">Grade </span>
      {grade}
    </span>
  )
}

const FLAG_STYLE: Record<Flag, string> = {
  missing: 'text-alert border-alert/40',
  late: 'text-warn border-warn/40',
  incomplete: 'text-warn border-warn/40',
  exempt: 'text-muted-foreground border-border',
  collected: 'text-now border-now/40',
  'not counted': 'text-faint border-border',
}

export function FlagChip({ flag }: { flag: Flag }) {
  return (
    <span className={cn('inline-flex h-5 items-center rounded-full border px-2 text-[11px] font-medium capitalize', FLAG_STYLE[flag])}>
      {flag}
    </span>
  )
}

export function CourseDot({ hue, className }: { hue: number | undefined; className?: string }) {
  return <span aria-hidden className={cn('inline-block size-2.5 shrink-0 rounded-full', className)} style={{ background: hueColor(hue) }} />
}
