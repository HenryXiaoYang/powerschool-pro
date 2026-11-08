import { cn } from '@/lib/utils'

/** The frosted-white selection bead. Its view-transition name lets it glide to the next selection. */
export function Pill({ name, className }: { name: string; className?: string }) {
  return <span aria-hidden className={cn('bead absolute inset-0 rounded-full bg-primary', className)} style={{ viewTransitionName: name }} />
}
