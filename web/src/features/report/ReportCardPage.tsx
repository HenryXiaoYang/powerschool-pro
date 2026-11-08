import { ArrowLeft, Printer } from 'lucide-react'
import { useRef, useState } from 'react'
import { Link, useParams } from 'react-router'
import { Panel } from '@/components/glass/Panel'
import { ErrorState, LoadingRows } from '@/components/glass/states'
import { Button } from '@/components/ui/button'
import { useReportCard, useReportCards } from '@/lib/queries'

/**
 * The school's own report card, rendered as sent (server-sanitized). The frame is
 * sandboxed without scripts; same-origin is allowed only so we can size and print it.
 */
export function ReportCardPage() {
  const { id = '' } = useParams()
  const { data: html, error, refetch, isPending } = useReportCard(id)
  const list = useReportCards(true)
  const name = list.data?.find((r) => r.id === id)?.name
  const frame = useRef<HTMLIFrameElement>(null)
  const [height, setHeight] = useState(1100)

  return (
    <div className="space-y-3 sm:space-y-4">
      <div className="flex items-center justify-between gap-3">
        <Link to="/report" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" /> Report
        </Link>
        <Button variant="secondary" className="glass rounded-full" disabled={!html} onClick={() => frame.current?.contentWindow?.print()}>
          <Printer /> Print or save PDF
        </Button>
      </div>
      {name && <h2 className="text-xl font-semibold tracking-tight">{name}</h2>}
      {isPending ? (
        <Panel><LoadingRows rows={8} /></Panel>
      ) : error ? (
        <Panel><ErrorState error={error} onRetry={() => void refetch()} /></Panel>
      ) : (
        <div className="glass overflow-x-auto rounded-xl p-2 sm:p-3">
          <iframe
            ref={frame}
            title={name ?? 'Report card'}
            srcDoc={html}
            sandbox="allow-same-origin allow-modals"
            className="block w-full min-w-[760px] rounded-lg bg-white"
            style={{ height }}
            onLoad={(e) => {
              const doc = e.currentTarget.contentDocument
              if (doc) setHeight(Math.max(400, Math.ceil(doc.body.getBoundingClientRect().height) + 16))
            }}
          />
        </div>
      )}
    </div>
  )
}
