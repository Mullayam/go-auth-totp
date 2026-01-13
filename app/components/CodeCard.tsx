import * as Clipboard from 'expo-clipboard';
import { LinearGradient } from 'expo-linear-gradient';
import * as OTPAuth from 'otpauth';
import React, { useEffect, useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { PremiumTheme } from '../constants/PremiumTheme';
import type { Account } from '../utils/storage';

interface Props {
    account: Account;
}

export const CodeCard = ({ account }: Props) => {
    const [code, setCode] = useState('');
    const [timeLeft, setTimeLeft] = useState(30);

    useEffect(() => {
        let totp = new OTPAuth.TOTP({
            secret: account.secret,
            algorithm: 'SHA1',
            digits: 6,
            period: 30,
        });

        const update = () => {
            setCode(totp.generate());
            const epoch = Math.round(new Date().getTime() / 1000.0);
            setTimeLeft(30 - (epoch % 30));
        };

        update();
        const interval = setInterval(update, 1000);
        return () => clearInterval(interval);
    }, [account]);

    const copyToClipboard = async () => {
        await Clipboard.setStringAsync(code);
    };

    const progressPercent = (timeLeft / 30) * 100;
    const isUrgent = timeLeft < 5;

    return (
        <TouchableOpacity onPress={copyToClipboard} activeOpacity={0.8} style={styles.touchable}>
            <LinearGradient
                colors={[PremiumTheme.colors.cardBackground, '#2A3648']}
                style={styles.card}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
            >
                <View style={styles.header}>
                    <Text style={styles.issuer}>{account.issuer || 'Enjoys'}</Text>
                    <Text style={styles.name} numberOfLines={1}>{account.name}</Text>
                </View>

                <View style={styles.mainContent}>
                    <Text style={styles.code}>{code.slice(0, 3)} {code.slice(3)}</Text>

                    <View style={styles.timerWrapper}>
                        <View style={[styles.timerTrack]}>
                            <View
                                style={[
                                    styles.timerFill,
                                    {
                                        width: `${progressPercent}%`,
                                        backgroundColor: isUrgent ? PremiumTheme.colors.error : PremiumTheme.colors.primary
                                    }
                                ]}
                            />
                        </View>
                        <Text style={styles.timerText}>{timeLeft}s</Text>
                    </View>
                </View>

                {/* Decorative Glow */}
                <View style={styles.activeIndicator} />
            </LinearGradient>
        </TouchableOpacity>
    );
};

const styles = StyleSheet.create({
    touchable: {
        marginVertical: 6,
        marginHorizontal: 16,
        ...PremiumTheme.shadows.card,
    },
    card: {
        borderRadius: PremiumTheme.borderRadius.l,
        padding: 20,
        borderWidth: 1,
        borderColor: PremiumTheme.colors.cardBorder,
    },
    header: {
        marginBottom: 12,
    },
    issuer: {
        fontSize: 11,
        color: PremiumTheme.colors.primary,
        fontWeight: '700',
        letterSpacing: 1,
        textTransform: 'uppercase',
        marginBottom: 4,
    },
    name: {
        fontSize: 14,
        color: PremiumTheme.colors.textSecondary,
        fontWeight: '500',
    },
    mainContent: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
    },
    code: {
        fontSize: 32,
        fontWeight: 'bold',
        color: PremiumTheme.colors.text,
        fontVariant: ['tabular-nums'],
        letterSpacing: 2,
        textShadowColor: 'rgba(99, 102, 241, 0.3)',
        textShadowOffset: { width: 0, height: 0 },
        textShadowRadius: 10,
    },
    timerWrapper: {
        alignItems: 'flex-end',
        width: 60,
    },
    timerTrack: {
        height: 4,
        width: '100%',
        backgroundColor: 'rgba(255,255,255,0.1)',
        borderRadius: 2,
        marginBottom: 6,
        overflow: 'hidden',
    },
    timerFill: {
        height: '100%',
        borderRadius: 2,
    },
    timerText: {
        fontSize: 10,
        color: PremiumTheme.colors.textSecondary,
        fontVariant: ['tabular-nums'],
    },
    activeIndicator: {
        position: 'absolute',
        top: 12,
        right: 12,
        width: 6,
        height: 6,
        borderRadius: 3,
        backgroundColor: PremiumTheme.colors.success,
        ...PremiumTheme.shadows.glow,
    }
});
