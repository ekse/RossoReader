import { describe, it, expect, beforeEach } from 'vitest'
import { useSidebar } from '../useSidebar'

describe('useSidebar Composable', () => {
  let sidebar: ReturnType<typeof useSidebar>

  beforeEach(() => {
    sidebar = useSidebar()
    // Reset state to true for predictable test conditions
    sidebar.open()
  })

  it('should initialize with isOpen equal to true', () => {
    expect(sidebar.isOpen.value).toBe(true)
  })

  it('should toggle isOpen state when toggle is called', () => {
    sidebar.toggle()
    expect(sidebar.isOpen.value).toBe(false)
    sidebar.toggle()
    expect(sidebar.isOpen.value).toBe(true)
  })

  it('should set isOpen to false when close is called', () => {
    sidebar.close()
    expect(sidebar.isOpen.value).toBe(false)
  })

  it('should set isOpen to true when open is called', () => {
    sidebar.close()
    expect(sidebar.isOpen.value).toBe(false)
    sidebar.open()
    expect(sidebar.isOpen.value).toBe(true)
  })
})
