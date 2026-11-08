import { LoaderCircle } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Navigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ApiError } from '@/lib/api'
import { signIn, useSession } from './session'

export function LoginPage() {
  const { signedIn } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [remember, setRemember] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (signedIn) return <Navigate to="/" replace />

  async function submit(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      await signIn(username.trim(), password, remember)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Sign-in failed. Try again.')
      setBusy(false)
    }
  }

  return (
    <main className="grid min-h-dvh place-items-center p-4">
      <form onSubmit={submit} className="glass settle w-full max-w-sm rounded-2xl p-6 sm:p-8" aria-describedby="login-note">
        <h1 className="text-3xl font-semibold tracking-tight">PowerSchool Pro</h1>
        <p className="mt-1 text-muted-foreground">Your PowerSchool, at a glance.</p>

        <div className="mt-8 space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="username">Username</Label>
            <Input id="username" autoComplete="username" autoCapitalize="none" spellCheck={false} required value={username} onChange={(e) => setUsername(e.target.value)} className="h-11" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="password">Password</Label>
            <Input id="password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} className="h-11" />
          </div>
          <div className="flex items-start gap-2.5">
            <Checkbox id="remember" checked={remember} onCheckedChange={(v) => setRemember(v === true)} className="mt-0.5" />
            <div className="grid gap-0.5">
              <Label htmlFor="remember">Remember me on this device</Label>
              <p className="text-xs text-muted-foreground">Off: you sign in again after closing this tab.</p>
            </div>
          </div>
        </div>

        {error && (
          <p role="alert" className="mt-5 rounded-md border border-alert/40 px-3 py-2 text-sm text-alert">
            {error}
          </p>
        )}

        <Button type="submit" disabled={busy} className="bead mt-6 h-11 w-full rounded-md text-[15px] font-semibold hover:bg-primary/90">
          {busy && <LoaderCircle className="animate-spin" />}
          {busy ? 'Signing in…' : 'Sign in'}
        </Button>

        <p id="login-note" className="mt-6 text-xs leading-relaxed text-muted-foreground">
          Use your school PowerSchool account. Your password is encrypted on this device and only sent to sign you in.
        </p>
      </form>
    </main>
  )
}
