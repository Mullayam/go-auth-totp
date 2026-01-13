import { Ionicons, MaterialIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { CameraType, CameraView, useCameraPermissions } from 'expo-camera';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, Dimensions, StatusBar, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { PremiumTheme } from '../constants/PremiumTheme';
import { saveAccount } from '../utils/storage';

const { width, height } = Dimensions.get('window');
const SCAN_SIZE = 280;

export default function App() {
    const [facing, setFacing] = useState<CameraType>('back');
    const [permission, requestPermission] = useCameraPermissions();
    const router = useRouter();
    const [scanned, setScanned] = useState(false);

    if (!permission) {
        return <View style={{ backgroundColor: PremiumTheme.colors.background, flex: 1 }} />;
    }

    if (!permission.granted) {
        return (
            <View style={[styles.container, { alignItems: 'center', justifyContent: 'center' }]}>
                <Text style={styles.message}>Camera permission required</Text>
                <TouchableOpacity style={styles.permissionBtn} onPress={requestPermission}>
                    <Text style={styles.permissionBtnText}>Grant Permission</Text>
                </TouchableOpacity>
            </View>
        );
    }

    const handleBarCodeScanned = async ({ type, data }: { type: string; data: string }) => {
        if (scanned) return;
        setScanned(true);

        try {
            if (!data.startsWith('otpauth://')) {
                Alert.alert('Invalid QR Code', 'This does not look like a valid authenticator code.', [
                    { text: 'Scan Again', onPress: () => setScanned(false) }
                ]);
                return;
            }

            const url = new URL(data);
            const secret = url.searchParams.get('secret');
            const issuer = url.searchParams.get('issuer') || 'Unknown';
            const label = decodeURIComponent(url.pathname.replace(/^\/\/totp\//, ''));
            const name = label.includes(':') ? label.split(':')[1] : label;

            if (!secret) {
                Alert.alert('Invalid Code', 'No secret key found in QR code.', [
                    { text: 'Scan Again', onPress: () => setScanned(false) }
                ]);
                return;
            }

            await saveAccount({
                name: name || 'Account',
                issuer: issuer,
                secret: secret,
                type: 'totp'
            });

            Alert.alert('Success', 'Account added successfully!', [
                { text: 'Done', onPress: () => router.back() }
            ]);

        } catch (e) {
            Alert.alert('Error', 'Failed to parse QR code.', [
                { text: 'Try Again', onPress: () => setScanned(false) }
            ]);
        }
    };

    function toggleCameraFacing() {
        setFacing(current => (current === 'back' ? 'front' : 'back'));
    }

    return (
        <View style={styles.container}>
            <StatusBar barStyle="light-content" />
            <CameraView
                style={styles.camera}
                facing={facing}
                onBarcodeScanned={scanned ? undefined : handleBarCodeScanned}
                barcodeScannerSettings={{
                    barcodeTypes: ["qr"],
                }}
            >
                <View style={styles.maskContainer}>
                    <View style={styles.maskRow} />
                    <View style={styles.maskCenter}>
                        <View style={styles.maskColumn} />
                        <View style={styles.scanWindow}>
                            <View style={[styles.corner, styles.tl]} />
                            <View style={[styles.corner, styles.tr]} />
                            <View style={[styles.corner, styles.bl]} />
                            <View style={[styles.corner, styles.br]} />
                        </View>
                        <View style={styles.maskColumn} />
                    </View>
                    <View style={styles.maskRow} />
                </View>

                <TouchableOpacity style={styles.closeBtn} onPress={() => router.back()}>
                    <BlurView intensity={20} tint="dark" style={styles.blurBtn}>
                        <Ionicons name="close" size={28} color="white" />
                    </BlurView>
                </TouchableOpacity>

                <View style={styles.controls}>
                    <Text style={styles.instruction}>Align QR code within the frame</Text>
                    <TouchableOpacity style={styles.flipBtn} onPress={toggleCameraFacing}>
                        <BlurView intensity={20} tint="dark" style={styles.blurBtn}>
                            <MaterialIcons name="flip-camera-ios" size={24} color="white" style={{ marginRight: 8 }} />
                            <Text style={styles.flipText}>Flip Camera</Text>
                        </BlurView>
                    </TouchableOpacity>
                </View>
            </CameraView>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: 'black',
    },
    message: {
        color: 'white',
        fontSize: 16,
        marginBottom: 20,
    },
    permissionBtn: {
        backgroundColor: PremiumTheme.colors.primary,
        paddingHorizontal: 20,
        paddingVertical: 12,
        borderRadius: 8,
    },
    permissionBtnText: {
        color: 'white',
        fontWeight: 'bold',
    },
    camera: {
        flex: 1,
    },
    maskContainer: {
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
    },
    maskRow: {
        flex: 1,
        backgroundColor: 'rgba(0,0,0,0.7)',
    },
    maskCenter: {
        height: SCAN_SIZE,
        flexDirection: 'row',
    },
    maskColumn: {
        flex: 1,
        backgroundColor: 'rgba(0,0,0,0.7)',
    },
    scanWindow: {
        width: SCAN_SIZE,
        height: SCAN_SIZE,
        backgroundColor: 'transparent',
        position: 'relative',
    },
    corner: {
        position: 'absolute',
        width: 20,
        height: 20,
        borderColor: PremiumTheme.colors.primary,
        borderWidth: 4,
    },
    tl: { top: 0, left: 0, borderRightWidth: 0, borderBottomWidth: 0 },
    tr: { top: 0, right: 0, borderLeftWidth: 0, borderBottomWidth: 0 },
    bl: { bottom: 0, left: 0, borderRightWidth: 0, borderTopWidth: 0 },
    br: { bottom: 0, right: 0, borderLeftWidth: 0, borderTopWidth: 0 },

    closeBtn: {
        position: 'absolute',
        top: 50,
        right: 20,
        overflow: 'hidden',
        borderRadius: 20,
    },
    controls: {
        position: 'absolute',
        bottom: 50,
        width: '100%',
        alignItems: 'center',
    },
    instruction: {
        color: 'white',
        fontSize: 14,
        marginBottom: 30,
        opacity: 0.8,
    },
    flipBtn: {
        borderRadius: 25,
        overflow: 'hidden',
    },
    blurBtn: {
        paddingHorizontal: 20,
        paddingVertical: 12,
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(255,255,255,0.1)',
    },
    flipText: {
        color: 'white',
        fontWeight: '600',
    }
});
