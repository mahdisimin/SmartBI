import { test, expect } from '@playwright/test'
import { API, collectConsoleErrors, dropUser, newPhone, seedUser, sql, uiLogin, uiLoginOk } from './helpers'

test.describe('auth', () => {
  test('register via UI -> success modal -> login -> dashboards list', async ({ page }) => {
    const errors = collectConsoleErrors(page)
    const phone = newPhone()
    await page.goto('/register')
    await page.fill('#userName', 'QA Register')
    await page.fill('#phoneNumber', phone)
    await page.fill('#password', 'qa-pass-123')
    await page.fill('#confirmPassword', 'qa-pass-123')
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.getByText('Registration Successful')).toBeVisible()
    const id = Number(sql(`SELECT ID FROM APP.[USER] WHERE PhoneNumber='${phone}'`))
    try {
      await expect(page.getByText(String(id), { exact: true })).toBeVisible()
      await page.getByRole('button', { name: 'Go to Login' }).click()
      await uiLogin(page, phone, 'qa-pass-123')
      await expect(page).toHaveURL('/')
      await expect(page.getByText('QA Register')).toBeVisible()
      await expect(page.getByText('No dashboard access')).toBeVisible()
      expect(errors, 'console errors').toEqual([])
    } finally {
      dropUser(id)
    }
  })

  test('register duplicate phone shows server error', async ({ page }) => {
    const u = await seedUser({ withSynops: false })
    try {
      await page.goto('/register')
      await page.fill('#userName', 'Dup')
      await page.fill('#phoneNumber', u.phone)
      await page.fill('#password', 'qa-pass-123')
      await page.fill('#confirmPassword', 'qa-pass-123')
      await page.getByRole('button', { name: 'Create account' }).click()
      await expect(page.getByText(/already registered/)).toBeVisible()
    } finally {
      dropUser(u.id)
    }
  })

  test('wrong password shows error and stays logged out', async ({ page }) => {
    const u = await seedUser({ withSynops: false })
    try {
      await uiLogin(page, u.phone, 'wrong-password')
      await expect(page.getByText('Phone number or password is incorrect.')).toBeVisible()
      await expect(page).toHaveURL('/login')
      const me = await page.request.get(`${API}/user/user_profile/me`)
      expect(me.status(), 'no session after failed login').toBe(401)
    } finally {
      dropUser(u.id)
    }
  })

  test('login phone with surrounding whitespace', async ({ page }) => {
    const u = await seedUser({ withSynops: false })
    try {
      await uiLogin(page, ` ${u.phone} `, u.password)
      await expect(page, 'register trims the phone but login does not').toHaveURL('/', { timeout: 5000 })
    } finally {
      dropUser(u.id)
    }
  })

  test('logged-out deep link redirects to /login', async ({ page }) => {
    await page.goto('/dashboards/synops')
    await expect(page).toHaveURL('/login')
  })

  test('logout clears session; back button does not restore it', async ({ page }) => {
    const u = await seedUser({ withSynops: true })
    try {
      await uiLogin(page, u.phone, u.password)
      await expect(page).toHaveURL('/')
      await page.getByRole('button', { name: 'Sign out' }).click()
      await expect(page).toHaveURL('/login')
      await page.goto('/dashboards/synops')
      await expect(page).toHaveURL('/login')
      const me = await page.request.get(`${API}/user/user_profile/me`)
      expect(me.status(), 'session revoked after logout').toBe(401)
    } finally {
      dropUser(u.id)
    }
  })

  test('SECURITY: forged session cookie / localStorage do not grant access', async ({ page, context }) => {
    await context.addCookies([{ name: 'smartbi_session', value: 'forged-token-abc', domain: 'localhost', path: '/' }])
    await page.goto('/login')
    await page.evaluate(() => localStorage.setItem('auth-storage', JSON.stringify({ state: { userId: 1 }, version: 0 })))
    await page.goto('/dashboards/synops')
    await expect(page).toHaveURL(/\/login/)
    expect(await page.locator('.kpi-card').count()).toBe(0)
  })

  test('SECURITY: export API is readable without any credentials', async ({ request }) => {
    const res = await request.get(`${API}/export/synops`)
    expect(res.status(), 'unauthenticated GET /export/synops').toBe(401)
  })
})

test.describe('session auth (batch: cookies + grants)', () => {
  test('session survives reload; deep link returns to original page after login', async ({ page }) => {
    const u = await seedUser({ withSynops: true })
    try {
      await page.goto('/dashboards/synops')
      await expect(page).toHaveURL(/\/login/)
      await uiLogin(page, u.phone, u.password)
      await expect(page, 'redirect back to the deep-linked page').toHaveURL('/dashboards/synops')
      await page.reload()
      await expect(page).toHaveURL('/dashboards/synops')
      await expect(page.locator('.kpi-card').first()).toBeVisible()
      const cookies = await page.context().cookies('http://localhost:8091')
      const c = cookies.find((x) => x.name === 'smartbi_session')
      expect(c?.httpOnly, 'session cookie HttpOnly').toBe(true)
      expect(await page.evaluate(() => document.cookie)).not.toContain('smartbi_session')
    } finally {
      dropUser(u.id)
    }
  })

  test('user with zero grants: empty dashboard list + real 403 screen', async ({ page }) => {
    const u = await seedUser({ withSynops: false })
    try {
      await uiLogin(page, u.phone, u.password)
      await expect(page).toHaveURL('/')
      await expect(page.getByText('No dashboard access')).toBeVisible()
      const exportCall = page.waitForResponse((r) => r.url().includes('/export/synops'))
      await page.goto('/dashboards/synops')
      expect((await exportCall).status()).toBe(403)
      await expect(page.getByText('شما به این داشبورد دسترسی ندارید')).toBeVisible()
      await expect(page, 'no redirect to login on 403').toHaveURL('/dashboards/synops')
      expect(await page.locator('.kpi-card').count()).toBe(0)
    } finally {
      dropUser(u.id)
    }
  })

  test('switching users in one browser does not leak the previous user data', async ({ page }) => {
    const a = await seedUser({ withSynops: true })
    const b = await seedUser({ withSynops: false })
    try {
      await uiLoginOk(page, a.phone, a.password)
      await expect(page.getByText('Synops Activity')).toBeVisible()
      await page.getByRole('button', { name: 'Sign out' }).click()
      await expect(page).toHaveURL('/login')
      await uiLoginOk(page, b.phone, b.password)
      await expect(page.getByText('No dashboard access')).toBeVisible()
      await expect(page.getByText('Synops Activity')).toHaveCount(0)
    } finally {
      dropUser(a.id); dropUser(b.id)
    }
  })

  test('session revoked server-side mid-use -> next API call sends user to login', async ({ page }) => {
    const u = await seedUser({ withSynops: true })
    try {
      await uiLoginOk(page, u.phone, u.password)
      await page.goto('/dashboards/synops')
      await expect(page.locator('.kpi-card').first()).toBeVisible()
      sql(`DELETE FROM APP.UserSession WHERE UserID = ${u.id}`)
      await page.locator('.rank-row').first().click() // triggers a refetch
      await expect(page).toHaveURL(/\/login/, { timeout: 10000 })
    } finally {
      dropUser(u.id)
    }
  })
})

test.describe('synops dashboard', () => {
  let u: { id: number; phone: string; password: string }
  test.beforeAll(async () => { u = await seedUser({ withSynops: true }) })
  test.afterAll(() => dropUser(u.id))

  test('card -> dashboard renders KPIs matching the API', async ({ page }) => {
    const errors = collectConsoleErrors(page)
    await uiLoginOk(page, u.phone, u.password)
    await page.getByText('Synops Activity').click()
    await expect(page).toHaveURL('/dashboards/synops')
    const res = await page.request.get(`${API}/export/synops`)
    expect(res.status()).toBe(200)
    const api = await res.json()
    const total = page.locator('.kpi-card').first().locator('.val')
    await expect(total).toHaveText(api.kpis.total_events.toLocaleString('en-US'))
    expect(errors, 'console errors').toEqual([])
  })

  test('cross-filter: click module -> chip + refetch; clear all restores', async ({ page }) => {
    await uiLoginOk(page, u.phone, u.password)
    await page.goto('/dashboards/synops')
    const total = page.locator('.kpi-card').first().locator('.val')
    await expect(total).not.toHaveText('')
    const before = await total.textContent()
    const row = page.locator('.rank-row').nth(1)
    const name = (await row.locator('.name').textContent())!.trim()
    const count = (await row.locator('.num').textContent())!.trim()
    await row.click()
    await expect(page.locator('.filter-chip')).toContainText(name)
    await expect(total).toHaveText(Number(count).toLocaleString('en-US'))
    await page.locator('.clear-all').click()
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    await expect(total).toHaveText(before!)
  })

  test('trend level toggle day/week/month', async ({ page }) => {
    await uiLoginOk(page, u.phone, u.password)
    await page.goto('/dashboards/synops')
    for (const lvl of ['هفته', 'ماه', 'روز']) {
      const btn = page.locator('.gran-toggle button', { hasText: lvl })
      await btn.click()
      await expect(btn).toHaveClass(/active/)
      await expect(page.locator('.panel h3 .cnt').first()).toContainText(lvl)
    }
    // BUG-23: month view shows Jalali month labels matching Jalali buckets
    await page.locator('.gran-toggle button', { hasText: 'ماه' }).click()
    const res = await page.request.get(`${API}/export/synops`)
    const months: { key: string }[] = (await res.json()).monthly_trend
    const names = ['فروردین','اردیبهشت','خرداد','تیر','مرداد','شهریور','مهر','آبان','آذر','دی','بهمن','اسفند']
    for (const m of months) {
      const [y, mm] = m.key.split('-')
      expect(Number(y), `month key ${m.key} is Jalali`).toBeGreaterThan(1390)
      expect(names[Number(mm) - 1], `month key ${m.key}`).toBeTruthy()
    }
    const labels: string[] = await page.evaluate(() => {
      // Chart.js keeps the category labels on the chart instance
      const w = window as unknown as { Chart?: { instances: Record<string, { data: { labels: string[] } }> } }
      const inst = w.Chart ? Object.values(w.Chart.instances) : []
      return inst.flatMap((c) => c.data.labels ?? [])
    })
    if (labels.length) {
      for (const m of months) {
        const [y, mm] = m.key.split('-')
        expect(labels, 'rendered month label').toContain(`${names[Number(mm) - 1]} ${y}`)
      }
    }
  })

  test('a11y: cross-filter rows are keyboard operable (Enter/Space)', async ({ page }) => {
    await uiLoginOk(page, u.phone, u.password)
    await page.goto('/dashboards/synops')
    const row = page.locator('.rank-row').first()
    await expect(row).toBeVisible()
    await expect(row).toHaveAttribute('role', 'button')
    await expect(row).toHaveAttribute('tabindex', '0')
    await expect(row).toHaveAttribute('aria-pressed', 'false')
    const name = (await row.locator('.name').textContent())!.trim()
    await row.focus()
    await expect(row).toBeFocused()
    await page.keyboard.press('Enter')
    await expect(page.locator('.filter-chip')).toContainText(name)
    await expect(row).toHaveAttribute('aria-pressed', 'true')
    await row.focus()
    await page.keyboard.press(' ')
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    // Space must not scroll the page
    // user table: plain rows + real toggle button in first cell (BUG-21)
    const tr = page.locator('table tbody tr').first()
    expect(await tr.getAttribute('role'), 'data <tr> must stay a plain row').toBeNull()
    const btn = tr.locator('td').first().locator('button')
    await expect(btn).toHaveAttribute('aria-pressed', 'false')
    await btn.focus()
    await page.keyboard.press('Enter')
    await expect(page.locator('.filter-chip')).toHaveCount(1)
    await expect(page.locator('.filter-chip')).toContainText('#')
    await expect(btn).toHaveAttribute('aria-pressed', 'true')
    await btn.focus()
    await page.keyboard.press(' ')
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    // mouse: button click toggles exactly once (no double toggle via row onClick)
    await btn.click()
    await expect(page.locator('.filter-chip')).toHaveCount(1)
    await btn.click()
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    // mouse: clicking elsewhere on the row still toggles once
    await tr.locator('td').nth(3).click()
    await expect(page.locator('.filter-chip')).toHaveCount(1)
    await expect(btn).toHaveAttribute('aria-pressed', 'true')
    await tr.locator('td').nth(3).click()
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    // screen reader tree: data rows are rows, with a pressed-state button in the first cell
    const snap = await page.locator('table').first().ariaSnapshot()
    const body = snap.split('rowgroup:')[2] ?? ''
    expect(body, 'tbody rows exposed as row, not button').toMatch(/^\s*- row /m)
    expect(body).toMatch(/- button "#\d+"/)
    await tr.screenshot({ path: 'test-results/bug21-row.png' })
    await btn.focus()
    await page.keyboard.press('Shift+Tab'); await page.keyboard.press('Tab')
    await page.locator('table').first().screenshot({ path: 'test-results/bug21-focus.png' })
    // reachable via Tab order (no filters active at this point)
    await expect(page.locator('.filter-chip')).toHaveCount(0)
    await page.locator('body').focus()
    let reached = false
    for (let i = 0; i < 60 && !reached; i++) {
      await page.keyboard.press('Tab')
      reached = await page.evaluate(() => document.activeElement?.classList.contains('rank-row') ?? false)
    }
    expect(reached, '.rank-row reachable with Tab').toBe(true)
  })

  test('backend error -> readable error state (no crash)', async ({ page }) => {
    await uiLoginOk(page, u.phone, u.password)
    await page.route('**/export/synops**', (r) => r.fulfill({ status: 500, body: '{"message":"boom"}', contentType: 'application/json' }))
    await page.goto('/dashboards/synops')
    await expect(page.getByText(/خطا در دریافت داده از سرور/)).toBeVisible({ timeout: 15000 })
  })

  test('mobile viewport: no horizontal overflow', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 800 })
    await uiLoginOk(page, u.phone, u.password)
    await page.goto('/dashboards/synops')
    await expect(page.locator('.kpi-card').first()).toBeVisible()
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)
    expect(overflow, 'horizontal overflow px at 375px').toBeLessThanOrEqual(0)
  })
})
