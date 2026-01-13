export const PremiumTheme = {
    colors: {
        background: '#0F172A', // Dark Slate Blue
        backgroundGradientStart: '#0F172A',
        backgroundGradientEnd: '#1E293B',

        cardBackground: '#1E293B',
        cardBorder: 'rgba(255, 255, 255, 0.1)',

        primary: '#6366F1', // Indigo
        primaryGradientStart: '#6366F1',
        primaryGradientEnd: '#8B5CF6', // Violet

        text: '#F8FAFC',
        textSecondary: '#94A3B8',

        success: '#10B981',
        error: '#EF4444',

        white: '#FFFFFF',
        black: '#000000',
    },
    spacing: {
        s: 8,
        m: 16,
        l: 24,
        xl: 32,
    },
    borderRadius: {
        m: 12,
        l: 16,
        xl: 24,
    },
    shadows: {
        card: {
            shadowColor: '#000',
            shadowOffset: { width: 0, height: 4 },
            shadowOpacity: 0.3,
            shadowRadius: 8,
            elevation: 6,
        },
        glow: {
            shadowColor: '#6366F1',
            shadowOffset: { width: 0, height: 0 },
            shadowOpacity: 0.4,
            shadowRadius: 10,
            elevation: 6,
        }
    }
};
