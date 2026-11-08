import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface PanelProps extends Omit<ComponentProps<'section'>, 'title'> {
  title?: ReactNode
  action?: ReactNode
}

/** A frosted tile. Titles are plain headings; the content carries the weight. */
export function Panel({ title, action, className, children, ...props }: PanelProps) {
  return (
    <section className={cn('glass rounded-xl p-3.5 sm:p-5', className)} {...props}>
      {(title || action) && (
        <header className="mb-2 flex min-h-8 sm:mb-3 items-center justify-between gap-3">
          {title && <h2 className="text-[15px] font-semibold tracking-tight">{title}</h2>}
          {action}
        </header>
      )}
      {children}
    </section>
  )
}
