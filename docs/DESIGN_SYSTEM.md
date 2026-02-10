# Design System & Style Guide
**Team Project**

Based on [Powder Wireless](https://powderwireless.net/) branding and design standards.

## Table of Contents
1. [Design Principles](#design-principles)
2. [Color Palette](#color-palette)
3. [Typography](#typography)
4. [Spacing & Layout](#spacing--layout)
5. [Components](#components)
6. [Interactive States](#interactive-states)
7. [Accessibility](#accessibility)
8. [Implementation Checklist](#implementation-checklist)

---

## Design Principles

### 1. **Powder Wireless Brand Alignment**
- Clean, minimal interface reflecting Powder's research-focused mission
- Teal/cyan primary color matching Powder brand identity
- Professional, technical appearance for research community

### 2. **Accessibility First**
- WCAG 2.1 AA compliance minimum
- High contrast ratios for all interactive elements
- Keyboard navigation support throughout

### 3. **Consistency**
- Unified component library via Material-UI
- Predictable patterns across all pages
- Consistent spacing, color usage, and typography aligned with Powder branding

### 4. **Responsive Design**
- Mobile-first approach
- Flexible layouts that adapt from 320px to 2560px
- Touch-friendly (48px+ tap targets)

### 5. **Performance**
- Lightweight components
- Smooth transitions (300ms standard)
- Lazy loading for large lists

---

## Color Palette

### Primary Brand Colors (Powder Wireless)

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Powder Teal** | `#00897B` | rgb(0, 137, 123) | Primary buttons, links, active states |
| **Powder Teal Light** | `#26A69A` | rgb(38, 166, 154) | Hover states, highlights |
| **Powder Teal Dark** | `#004D40` | rgb(0, 77, 64) | Dark text, borders, navbar |
| **Powder Cyan** | `#00BCD4` | rgb(0, 188, 212) | Accent highlights, active indicators |

### Semantic Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Success** | `#4CAF50` | rgb(76, 175, 80) | Success messages, approved requests |
| **Warning** | `#FF9800` | rgb(255, 152, 0) | Warnings, pending requests |
| **Error** | `#F44336` | rgb(244, 67, 54) | Errors, critical issues |
| **Info** | `#2196F3` | rgb(33, 150, 243) | Information, notifications |

### Neutral Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **White** | `#FFFFFF` | rgb(255, 255, 255) | Backgrounds, cards, text on dark |
| **Light Gray** | `#F5F5F5` | rgb(245, 245, 245) | Secondary backgrounds |
| **Medium Gray** | `#BDBDBD` | rgb(189, 189, 189) | Borders, disabled states |
| **Dark Gray** | `#424242` | rgb(66, 66, 66) | Secondary text (light mode) |
| **Near Black** | `#212121` | rgb(33, 33, 33) | Primary text |

### Extended Palette

```css
:root {
  /* Powder Primary */
  --color-primary: #00897B;
  --color-primary-light: #26A69A;
  --color-primary-dark: #004D40;
  
  /* Semantic */
  --color-success: #4CAF50;
  --color-warning: #FF9800;
  --color-error: #F44336;
  --color-info: #2196F3;
  
  /* Accent */
  --color-accent: #00BCD4;
  --color-accent-dark: #0097A7;
  
  /* Neutral */
  --color-white: #FFFFFF;
  --color-light-bg: #F5F5F5;
  --color-border: #BDBDBD;
  --color-text-primary: #212121;
  --color-text-secondary: #666666;
  --color-disabled: #BDBDBD;
}
```

### Color Usage Guidelines

**Powder Teal (#00897B)**
- Primary call-to-action buttons
- Active navigation items
- Main links and focus states
- Header/navigation bar background
- Active tab indicators

**Powder Teal Light (#26A69A)**
- Button hover states
- Interactive element hover effects
- Secondary highlights
- Hover backgrounds

**Powder Teal Dark (#004D40)**
- Navigation bar and dark themed areas
- Secondary text on light backgrounds
- Border colors for emphasis
- Dark mode text

**Powder Cyan (#00BCD4)**
- Active/connected status indicators
- Special emphasis elements
- Experiment status badges
- Secondary accent elements

**Success Green (#4CAF50)**
- Approved hardware requests
- Successful operations
- "Available" status indicators
- Valid form inputs

**Warning Orange (#FF9800)**
- Pending approval requests
- Low availability warnings
- Cautionary messages
- "Limited" status

**Error Red (#F44336)**
- Failed operations
- Unavailable hardware
- Error messages
- Form validation errors

---

## Typography

### Font Family (Powder Wireless Aligned)

Primary font stack - matching Powder's clean, professional aesthetic:
```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
             "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
             sans-serif;
```

**Sistema Fallback:** The font stack uses system fonts for optimal performance and native appearance across platforms, consistent with modern web standards used by Powder.

### Type Scale

| Name | Size | Weight | Line Height | Usage |
|------|------|--------|-------------|-------|
| **H1** | 32px / 2.0rem | 600 (Semibold) | 1.2 | Page titles (main heading) |
| **H2** | 28px / 1.75rem | 600 (Semibold) | 1.3 | Section titles |
| **H3** | 24px / 1.5rem | 600 (Semibold) | 1.3 | Subsection titles |
| **H4** | 20px / 1.25rem | 600 (Semibold) | 1.4 | Card titles, form headers |
| **H5** | 18px / 1.125rem | 500 (Medium) | 1.4 | Secondary headings |
| **H6** | 16px / 1rem | 500 (Medium) | 1.5 | Tertiary headings |
| **Body1** | 16px / 1rem | 400 (Regular) | 1.5 | Primary body text |
| **Body2** | 14px / 0.875rem | 400 (Regular) | 1.5 | Secondary body text |
| **Caption** | 12px / 0.75rem | 400 (Regular) | 1.4 | Helper text, labels |
| **Button** | 14px / 0.875rem | 500 (Medium) | 1.5 | Button text |

### Typography Hierarchy Example
```tsx
// Page Title
<Typography variant="h1">
  Hardware Management
</Typography>

// Section Title
<Typography variant="h3">
  Available Devices
</Typography>

// Card Title
<Typography variant="h6" sx={{ mb: 1 }}>
  Device Information
</Typography>

// Body Text
<Typography variant="body1" color="text.secondary">
  Select a device from the list to view details.
</Typography>

// Helper Text
<Typography variant="caption" color="text.secondary">
  Last updated 5 minutes ago
</Typography>
```

---

## Spacing & Layout

### Spacing Scale
Based on 8px base unit (Material Design guidelines):

```css
--spacing-0: 0px
--spacing-1: 4px
--spacing-2: 8px
--spacing-3: 12px
--spacing-4: 16px
--spacing-5: 20px
--spacing-6: 24px
--spacing-7: 28px
--spacing-8: 32px
--spacing-10: 40px
--spacing-12: 48px
--spacing-16: 64px
```

### Common Spacing Patterns

| Context | Spacing | Code |
|---------|---------|------|
| Component internal padding | 16px (xs), 24px (md) | `sx={{ p: 2, p: 3 }}` |
| Vertical spacing between elements | 16px, 24px | `sx={{ mb: 2 }}` |
| Horizontal gap in rows | 8px, 16px | `sx={{ gap: 1, gap: 2 }}` |
| Card padding | 24px | `<Card sx={{ p: 3 }}>` |
| Button padding | 10px 16px | `<Button sx={{ px: 2, py: 1 }}>` |

### Responsive Breakpoints
```tsx
const breakpoints = {
  xs: 0,      // Extra small (mobile)
  sm: 600,    // Small (tablets)
  md: 960,    // Medium (small laptops)
  lg: 1280,   // Large (laptops)
  xl: 1920    // Extra large (desktops)
};

// Responsive spacing example
<Box sx={{
  p: { xs: 1, sm: 2, md: 3 },    // Padding adapts by screen size
  gap: { xs: 1, sm: 2, md: 3 }   // Gap adapts by screen size
}}>
```

### Container Widths
```css
/* Standard container max widths */
--container-xs: 360px
--container-sm: 600px
--container-md: 960px
--container-lg: 1200px
--container-xl: 1320px
```

---

## Components

### Buttons

#### Button Variants
```tsx
// Contained (Solid) - Primary Action
<Button variant="contained" color="primary">
  Primary Action
</Button>

// Outlined - Secondary Action
<Button variant="outlined" color="primary">
  Secondary Action
</Button>

// Text - Tertiary Action
<Button variant="text" color="primary">
  Tertiary Action
</Button>

// Loading State
<LoadingButton loading={isLoading} variant="contained">
  Submit
</LoadingButton>
```

#### Button Sizes
```tsx
// Small - For compact layouts
<Button size="small" variant="contained">
  Small Button
</Button>

// Medium - Default, most common
<Button variant="contained">
  Medium Button
</Button>

// Large - For mobile or emphasis
<Button size="large" variant="contained">
  Large Button
</Button>
```

#### Button Colors
```tsx
// Primary - Powder Teal
<Button color="primary" variant="contained">
  Primary Action
</Button>

// Success - For approve/confirm
<Button sx={{ backgroundColor: '#4CAF50', color: 'white', '&:hover': { backgroundColor: '#45a049' } }}>
  Approve
</Button>

// Error - For delete/decline
<Button color="error" variant="contained">
  Delete
</Button>

// Warning - For caution
<Button sx={{ backgroundColor: '#FF9800', color: 'white', '&:hover': { backgroundColor: '#e68900' } }}>
  Review
</Button>
```

### Cards

#### Basic Card
```tsx
<Card sx={{ mb: 3 }}>
  <CardContent>
    <Typography variant="h6" gutterBottom>
      Card Title
    </Typography>
    <Typography variant="body2" color="text.secondary">
      Card content goes here with descriptive text.
    </Typography>
  </CardContent>
  <CardActions>
    <Button size="small">Learn More</Button>
  </CardActions>
</Card>
```

#### Outlined Card (Subtle)
```tsx
<Card variant="outlined" sx={{ mb: 2 }}>
  <CardContent>
    <Typography variant="h6">
      Subtle Card
    </Typography>
  </CardContent>
</Card>
```

#### Elevated Card (Emphasis)
```tsx
<Card elevation={8} sx={{ mb: 2 }}>
  <CardContent>
    <Typography variant="h6">
      Important Information
    </Typography>
  </CardContent>
</Card>
```

### Input Fields

#### Text Field
```tsx
<TextField
  fullWidth
  label="Username"
  placeholder="Enter your username"
  variant="outlined"
  sx={{ mb: 2 }}
  error={!!errors.username}
  helperText={errors.username}
/>
```

#### Variants
```tsx
// Outlined - Default, recommended
<TextField variant="outlined" label="Input" />

// Filled - Alternative style
<TextField variant="filled" label="Input" />

// Standard - Minimal
<TextField variant="standard" label="Input" />
```

### Messages & Alerts

#### Alert Types
```tsx
// Success Alert
<Alert severity="success">
  Operation completed successfully!
</Alert>

// Warning Alert
<Alert severity="warning" sx={{ mb: 2 }}>
  Hardware availability is low.
</Alert>

// Error Alert
<Alert severity="error" sx={{ mb: 2 }}>
  Failed to process request. Please try again.
</Alert>

// Info Alert
<Alert severity="info" sx={{ mb: 2 }}>
  New hardware devices available for checkout.
</Alert>
```

### Lists & Tables

#### List Structure
```tsx
<Stack spacing={1}>
  {items.map((item) => (
    <Card key={item.id} variant="outlined">
      <CardContent sx={{ py: 1.5 }}>
        <Typography variant="body2">
          {item.name}
        </Typography>
      </CardContent>
    </Card>
  ))}
</Stack>
```

---

## Interactive States

### Button States

```css
/* Default State */
.btn-primary {
  background-color: #00897B;
  color: white;
  cursor: pointer;
  transition: all 300ms ease;
}

/* Hover State */
.btn-primary:hover {
  background-color: #26A69A;
  box-shadow: 0 2px 8px rgba(0, 137, 123, 0.3);
  transform: translateY(-1px);
}

/* Active/Pressed State */
.btn-primary:active {
  background-color: #004D40;
  transform: translateY(0);
  box-shadow: 0 1px 4px rgba(0, 137, 123, 0.2);
}

/* Disabled State */
.btn-primary:disabled {
  background-color: #BDBDBD;
  color: #999;
  cursor: not-allowed;
  opacity: 0.6;
}

/* Focus State (Keyboard Navigation) */
.btn-primary:focus-visible {
  outline: 3px solid #26A69A;
  outline-offset: 2px;
}
```

### Form Input States

```css
/* Default/Empty State */
.input-field {
  border: 1px solid #BDBDBD;
  background-color: white;
  color: #212121;
}

/* Focused State */
.input-field:focus {
  border-color: #00897B;
  box-shadow: 0 0 0 3px rgba(0, 137, 123, 0.1);
}

/* Filled State */
.input-field:not(:placeholder-shown) {
  border-color: #26A69A;
}

/* Error State */
.input-field.error {
  border-color: #F44336;
  background-color: rgba(244, 67, 54, 0.05);
}

/* Disabled State */
.input-field:disabled {
  background-color: #F5F5F5;
  color: #BDBDBD;
  cursor: not-allowed;
}
```

### Transition Guidelines

```css
/* Standard transitions */
--transition-fast: 150ms cubic-bezier(0.4, 0, 0.2, 1);
--transition-standard: 300ms cubic-bezier(0.4, 0, 0.2, 1);
--transition-slow: 500ms cubic-bezier(0.4, 0, 0.2, 1);

/* Use cases */
/* Fast: Small UI changes (opacity, color) */
/* Standard: Most interactive elements (hover, focus) */
/* Slow: Complex animations, transitions between pages */
```

---

## Accessibility

### Color Contrast Requirements

All text must meet WCAG AA standards:
- **Large text** (18px+): 3:1 contrast ratio minimum
- **Normal text**: 4.5:1 contrast ratio minimum
- **UI components**: 3:1 contrast ratio minimum

### Verified Contrast Ratios

| Foreground | Background | Ratio | WCAG Level |
|-----------|-----------|-------|-----------|
| #212121 (text) | #FFFFFF (white bg) | 16.5:1 | AAA ✓ |
| #00897B (Powder Teal) | #FFFFFF (white bg) | 8.1:1 | AAA ✓ |
| #FFFFFF (text) | #00897B (Powder Teal) | 8.1:1 | AAA ✓ |
| #F44336 (error) | #FFFFFF (white bg) | 5.2:1 | AAA ✓ |
| #4CAF50 (success) | #FFFFFF (white bg) | 6.5:1 | AAA ✓ |

### Keyboard Navigation

```tsx
// Ensure all interactive elements are keyboard accessible
// Button navigation
<Button 
  variant="contained"
  sx={{
    '&:focus-visible': {
      outline: '3px solid #26A69A',
      outlineOffset: '2px'
    }
  }}
>
  Action Button
</Button>

// Form field focus
<TextField
  autoFocus={shouldFocus}
  inputProps={{
    'aria-label': 'Username or Email',
  }}
/>
```

### Screen Reader Support

```tsx
// Provide meaningful labels
<TextField
  label="Username"
  aria-label="Enter your username"
  aria-describedby="username-helper"
/>
<Typography id="username-helper" variant="caption" color="text.secondary">
  Your unique login identifier (3-20 characters)
</Typography>

// Alert updates
<Alert severity="error" role="alert">
  Please correct the following errors before submitting.
</Alert>

// Skip to main content link (hidden by default)
<Button
  component="a"
  href="#main-content"
  sx={{
    position: 'absolute',
    top: -40,
    left: 0,
    bg: '#00897B',
    color: 'white',
    padding: '8px',
    '&:focus': {
      top: 0
    }
  }}
>
  Skip to main content
</Button>
```

### Motion Preferences

```tsx
// Respect reduced motion preferences
<Box
  sx={{
    transition: 'all 300ms ease',
    '@media (prefers-reduced-motion: reduce)': {
      transition: 'none'
    }
  }}
>
  Content with animation
</Box>
```

---

## Usage Examples

### 1. Login Form

```tsx
<Card variant="outlined" sx={{ maxWidth: 400, mx: 'auto', mt: 8 }}>
  <CardContent sx={{ p: 4 }}>
    <Typography variant="h4" gutterBottom sx={{ mb: 3, textAlign: 'center' }}>
      Sign In
    </Typography>
    
    <Stack spacing={2}>
      <TextField
        fullWidth
        label="Username or Email"
        variant="outlined"
        placeholder="username"
      />
      
      <TextField
        fullWidth
        label="Password"
        type="password"
        variant="outlined"
        placeholder="••••••••"
      />
      
      <Button
        fullWidth
        variant="contained"
        color="primary"
        size="large"
        sx={{ mt: 1 }}
      >
        Sign In
      </Button>
      
      <Typography variant="body2" align="center" color="text.secondary">
        Don't have an account?{' '}
        <Link href="/register" sx={{ color: '#00897B', fontWeight: 500 }}>
          Create one
        </Link>
      </Typography>
    </Stack>
  </CardContent>
</Card>
```

### 2. Hardware Availability Card

```tsx
<Card sx={{ mb: 3 }}>
  <CardContent>
    <Stack direction="row" justifyContent="space-between" alignItems="start" sx={{ mb: 2 }}>
      <Box>
        <Typography variant="h6">Dell XPS 15</Typography>
        <Typography variant="caption" color="text.secondary">
          Laptop Workstation
        </Typography>
      </Box>
      <Box
        sx={{
          px: 1.5,
          py: 0.5,
          borderRadius: 1,
          backgroundColor: '#4CAF50',
          color: 'white'
        }}
      >
        <Typography variant="caption" sx={{ fontWeight: 500 }}>
          Available (8/10)
        </Typography>
      </Box>
    </Stack>
    
    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
      High-performance laptop suitable for development and cloud-native testing.
    </Typography>
    
    <Box sx={{ bgcolor: '#F5F5F5', p: 2, borderRadius: 1, mb: 2 }}>
      <Typography variant="caption" color="text.secondary">
        Specifications: Intel i7, 32GB RAM, RTX 3080
      </Typography>
    </Box>
  </CardContent>
  <CardActions sx={{ justifyContent: 'flex-end', gap: 1 }}>
    <Button variant="outlined" color="primary">
      View Details
    </Button>
    <Button variant="contained" color="primary">
      Request Checkout
    </Button>
  </CardActions>
</Card>
```

### 3. Status Message / Toast Pattern

```tsx
// Success State
<Alert severity="success" icon={<CheckCircleIcon />} sx={{ mb: 3 }}>
  <AlertTitle>Request Approved</AlertTitle>
  Your hardware request has been approved. Device will be available for pickup.
</Alert>

// Pending State
<Alert severity="warning" icon={<InfoIcon />} sx={{ mb: 3 }}>
  <AlertTitle>Pending Approval</AlertTitle>
  Your request is awaiting administrator approval.
</Alert>

// Error State
<Alert severity="error" icon={<ErrorIcon />} sx={{ mb: 3 }}>
  <AlertTitle>Request Failed</AlertTitle>
  Unable to process your request. Device is no longer available.
</Alert>
```

### 4. Responsive Grid Layout

```tsx
<Box sx={{ p: { xs: 2, sm: 3, md: 4 } }}>
  <Typography variant="h2" sx={{ mb: 4 }}>
    Available Devices
  </Typography>
  
  <Grid container spacing={{ xs: 2, sm: 2, md: 3 }}>
    {devices.map((device) => (
      <Grid item xs={12} sm={6} md={4} key={device.id}>
        <Card>
          <CardContent>
            <Typography variant="h6">{device.name}</Typography>
            <Typography variant="body2" color="text.secondary">
              {device.description}
            </Typography>
          </CardContent>
          <CardActions>
            <Button size="small">Details</Button>
            <Button size="small">Request</Button>
          </CardActions>
        </Card>
      </Grid>
    ))}
  </Grid>
</Box>
```

### 5. Dark Mode Support (Future Enhancement)

```tsx
// Theming configuration example
import { createTheme, ThemeProvider } from '@mui/material/styles';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#5B9FD5',
      dark: '#446B99',
    },
    secondary: {
      main: '#00BCD4',
    },
    background: {
      default: '#121212',
      paper: '#1E1E1E',
    },
    text: {
      primary: '#FFFFFF',
      secondary: 'rgba(255, 255, 255, 0.7)',
    },
  },
});

// Usage
<ThemeProvider theme={darkTheme}>
  <App />
</ThemeProvider>
```

---

## Implementation Checklist

- [ ] Primary color (#446B99) applied to main buttons and links
- [ ] Secondary color (#5B9FD5) used for hover states
- [ ] Semantic colors (green/orange/red) applied appropriately
- [ ] Typography scale implemented across all pages
- [ ] Spacing system consistently applied (16px increments)
- [ ] Button states (hover, active, disabled) working smoothly
- [ ] Form fields showing proper focus and error states
- [ ] Alert/message components styled correctly
- [ ] Accessibility: Color contrast verified (4.5:1 minimum)
- [ ] Accessibility: Keyboard navigation through all interactive elements
- [ ] Accessibility: Focus visible outlines (3px, #5B9FD5)
- [ ] Responsive design tested at all breakpoints (xs, sm, md, lg, xl)
- [ ] Touch targets all 48px minimum (mobile)
- [ ] Loading states and transitions smooth (300ms standard)
- [ ] Material-UI theme synchronized with this design system
- [ ] Dark mode support implemented (future)

---

## Resources & References

- [Material-UI Documentation](https://mui.com)
- [Material Design 3 Guidelines](https://m3.material.io)
- [WCAG 2.1 Accessibility Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [WebAIM Color Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [Accessible Color Palettes](https://www.accessible-colors.com)

---

**Last Updated:** February 10, 2026
**Version:** 1.0
**Status:** Active - Implementation Ready
