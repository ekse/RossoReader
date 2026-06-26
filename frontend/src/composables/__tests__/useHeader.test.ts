import { describe, it, expect, beforeEach } from 'vitest'
import { useHeader } from '../useHeader'

describe('useHeader Composable', () => {
  let header: ReturnType<typeof useHeader>

  beforeEach(() => {
    header = useHeader()
    // Reset state for predictable test conditions
    header.isHeaderVisible.value = true
    header.lastScrollTop.value = 0
  })

  it('should initialize with defaults', () => {
    expect(header.isHeaderVisible.value).toBe(true)
    expect(header.lastScrollTop.value).toBe(0)
  })

  it('should allow modifying states', () => {
    header.isHeaderVisible.value = false
    header.lastScrollTop.value = 100

    expect(header.isHeaderVisible.value).toBe(false)
    expect(header.lastScrollTop.value).toBe(100)
  })

  it('should share state globally across separate calls', () => {
    const header2 = useHeader()
    header.isHeaderVisible.value = false
    header.lastScrollTop.value = 250

    expect(header2.isHeaderVisible.value).toBe(false)
    expect(header2.lastScrollTop.value).toBe(250)
  })

  it('should handle scroll events on desktop and mobile', () => {
    const originalInnerWidth = window.innerWidth

    // Desktop: should always keep header visible
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 })
    const target = document.createElement('div')
    target.scrollTop = 100
    const event = new Event('scroll')
    Object.defineProperty(event, 'target', { value: target })

    header.isHeaderVisible.value = false
    header.handleScroll(event)
    expect(header.isHeaderVisible.value).toBe(true)

    // Mobile: scroll down past threshold
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 })
    target.scrollTop = 100
    header.handleScroll(event)
    expect(header.isHeaderVisible.value).toBe(false)
    expect(header.lastScrollTop.value).toBe(100)

    // Mobile: scroll up
    target.scrollTop = 50
    header.handleScroll(event)
    expect(header.isHeaderVisible.value).toBe(true)
    expect(header.lastScrollTop.value).toBe(50)

    // Cleanup
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalInnerWidth })
  })
})
