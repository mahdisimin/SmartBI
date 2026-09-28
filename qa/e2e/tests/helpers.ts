import { execSync } from 'node:child_process'
import type { Page } from '@playwright/test'

export const API = 'http://localhost:8091'

export function sql(q: string): string {
  return execSync(`sqlcmd -S . -E -d SMARTBI -I -b -W -h -1 -Q "SET NOCOUNT ON; ${q}"`, { encoding: 'utf8' }).trim()
}

export function newPhone(): string {
  return '09' + String(Date.now()).slice(-9)
}

/** Registers a user through the API and grants it the Synops dashboard link. */
export async function seedUser(opts: { withSynops: boolean }) {
  const phone = newPhone()
  const password = 'qa-pass-123'
  const res = await fetch(`${API}/user/register`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_name: 'QA E2E', phone_number: phone, password }),
  })
  const body = await res.json()
  if (!res.ok || typeof body.user_id !== "number") throw new Error(`seed register failed: ${res.status} ${JSON.stringify(body)}`)
  const UserId: number = body.user_id
  if (opts.withSynops) sql(`INSERT INTO APP.User_WebAppLink (UserID, LinkID) VALUES (${UserId}, 3)`)
  return { id: UserId as number, phone, password }
}

export function dropUser(id: number) {
  sql(`DELETE FROM APP.UserSession WHERE UserID = ${id}; DELETE FROM APP.User_WebAppLink WHERE UserID = ${id}; DELETE FROM APP.[USER] WHERE ID = ${id}`)
}

export function collectConsoleErrors(page: Page): string[] {
  const errors: string[] = []
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()) })
  page.on('pageerror', (e) => errors.push(String(e)))
  return errors
}

export async function uiLogin(page: Page, phone: string, password: string) {
  await page.goto('/login')
  await page.fill('#phone', phone)
  await page.fill('#password', password)
  await page.getByRole('button', { name: 'Sign in' }).click()
}

/** Logs in and waits until the app has landed on the dashboards list. */
export async function uiLoginOk(page: Page, phone: string, password: string) {
  await uiLogin(page, phone, password)
  await page.waitForURL('http://localhost:5173/')
}
