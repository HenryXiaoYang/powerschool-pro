import { LogOut, RotateCcw } from 'lucide-react'
import { useId, type ReactNode } from 'react'
import { Panel } from '@/components/glass/Panel'
import { Button } from '@/components/ui/button'
import { signOut } from '@/features/auth/session'
import { DEFAULTS, setAppearance, useAppearance } from '@/lib/appearance'
import { cn } from '@/lib/utils'

export function SettingsPage() {
  const a = useAppearance()
  const changed = (Object.keys(DEFAULTS) as (keyof typeof DEFAULTS)[]).some((k) => a[k] !== DEFAULTS[k])
  return (
    <div className="mx-auto grid max-w-2xl gap-3 sm:gap-4">
      <Panel
        title="Appearance"
        action={
          <Button variant="ghost" size="sm" className="h-8 gap-1.5 text-muted-foreground" disabled={!changed} onClick={() => setAppearance(DEFAULTS)}>
            <RotateCcw className="size-3.5" /> Reset
          </Button>
        }
      >
        <Preview />
        <div className="mt-4 grid gap-5">
          <Slider
            label="Glass opacity"
            hint="Lower shows more of the backdrop; higher is easier to read."
            value={a.opacity}
            min={20}
            max={200}
            step={5}
            format={(v) => `${v}%`}
            onChange={(opacity) => setAppearance({ opacity })}
          />
          <Slider
            label="Blur"
            hint="Heavy blur looks smoother but costs battery. Off is the fastest."
            value={a.blur}
            min={0}
            max={40}
            step={2}
            format={(v) => (v === 0 ? 'Off' : `${v} px`)}
            onChange={(blur) => setAppearance({ blur })}
          />
          <Switch label="Liquid glass edges" hint="A bright rim and sheen on every pane, like glass catching the light." checked={a.liquid} onChange={(liquid) => setAppearance({ liquid })} />
          <Switch label="Calm mode" hint="Keeps the backdrop still and skips pane animations. Easier on older phones." checked={a.calm} onChange={(calm) => setAppearance({ calm })} />
        </div>
      </Panel>
      <Panel title="Account">
        <Button variant="secondary" className="h-11 w-full justify-start gap-2 px-3 text-[15px]" onClick={() => void signOut()}>
          <LogOut className="size-4" /> Sign out
        </Button>
      </Panel>
    </div>
  )
}

/** A pane over loud colour, so every slider has an instant, honest effect. */
function Preview() {
  return (
    <div className="relative grid h-28 place-items-center overflow-hidden rounded-lg" aria-hidden>
      <div className="absolute inset-0 bg-[conic-gradient(from_20deg,oklch(0.75_0.15_250),oklch(0.8_0.14_190),oklch(0.78_0.15_320),oklch(0.75_0.15_250))]" />
      <div className="glass relative rounded-full px-6 py-3 text-sm font-medium">Glass looks like this</div>
    </div>
  )
}

function Slider({ label, hint, value, min, max, step, format, onChange }: { label: string; hint: string; value: number; min: number; max: number; step: number; format: (v: number) => string; onChange: (v: number) => void }) {
  const id = useId()
  return (
    <div>
      <div className="flex items-baseline justify-between gap-3">
        <label htmlFor={id} className="font-medium">{label}</label>
        <output htmlFor={id} className="tabular text-sm text-muted-foreground">{format(value)}</output>
      </div>
      <input
        id={id}
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(e.currentTarget.valueAsNumber)}
        className="mt-2 block h-8 w-full cursor-pointer accent-[var(--now)]"
        aria-describedby={`${id}-hint`}
      />
      <p id={`${id}-hint`} className="text-xs text-muted-foreground">{hint}</p>
    </div>
  )
}

function Switch({ label, hint, checked, onChange }: { label: ReactNode; hint: string; checked: boolean; onChange: (v: boolean) => void }) {
  const id = useId()
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <label htmlFor={id} className="font-medium">{label}</label>
        <p className="text-xs text-muted-foreground">{hint}</p>
      </div>
      <button
        id={id}
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={cn('relative h-8 w-14 shrink-0 rounded-full border transition-colors duration-200', checked ? 'border-transparent bg-now' : 'border-border bg-muted')}
      >
        <span className={cn('absolute top-0.5 left-0.5 size-7 rounded-full bg-white shadow transition-transform duration-200', checked && 'translate-x-6')} />
      </button>
    </div>
  )
}
