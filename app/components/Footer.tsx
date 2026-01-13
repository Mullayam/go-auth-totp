import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { PremiumTheme } from '../constants/PremiumTheme';

export const Footer = () => {
    return (
        <View style={styles.container}>
            <Text style={styles.text}>Made with ❤️ by Enjoys</Text>
        </View>
    );
};

const styles = StyleSheet.create({
    container: {
        padding: 24,
        alignItems: 'center',
        justifyContent: 'center',
        opacity: 0.7,
    },
    text: {
        color: PremiumTheme.colors.textSecondary,
        fontSize: 12,
        fontWeight: '500',
        letterSpacing: 0.5,
    },
});
