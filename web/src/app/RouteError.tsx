import { useRouteError } from 'react-router'
import { Panel } from '@/components/glass/Panel'
import { ErrorState } from '@/components/glass/states'

/** A page that crashed while rendering: say so in the product's words and offer a reload. */
export function RouteError() {
  console.error(useRouteError())
  return (
    <Panel className="mx-auto w-full max-w-md">
      <ErrorState error={new Error('This page hit a problem. Reloading usually fixes it.')} onRetry={() => location.reload()} />
    </Panel>
  )
}
