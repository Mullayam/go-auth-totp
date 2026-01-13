import { CameraType, CameraView, useCameraPermissions } from 'expo-camera';
import * as Linking from 'expo-linking';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, Button, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { saveAccount } from '../utils/storage';

export default function App() {
    const [facing, setFacing] = useState<CameraType>('back');
    const [permission, requestPermission] = useCameraPermissions();
    const router = useRouter();
    const [scanned, setScanned] = useState(false);

    if (!permission) {
        // Camera permissions are still loading.
        return <View />;
    }

    if (!permission.granted) {
        // Camera permissions are not granted yet.
        return (
            <View style={styles.container}>
                <Text style={styles.message}>We need your permission to show the camera</Text>
                <Button onPress={requestPermission} title="grant permission" />
            </View>
        );
    }

    const handleBarCodeScanned = async ({ type, data }: { type: string; data: string }) => {
        if (scanned) return;
        setScanned(true);

        try {
            // Expected format: otpauth://totp/Issuer:Account?secret=...&issuer=...
            const parsed = Linking.parse(data);

            // Basic validation logic since Linking.parse might not handle otpauth scheme perfectly if not configured
            // We can manually parse if needed, but let's try to extract basic info

            if (!data.startsWith('otpauth://')) {
                Alert.alert('Invalid QR Code', 'This does not look like a valid authenticator code.', [
                    { text: 'OK', onPress: () => setScanned(false) }
                ]);
                return;
            }

            const url = new URL(data);
            const secret = url.searchParams.get('secret');
            const issuer = url.searchParams.get('issuer') || 'Unknown';
            // Path usually contains "Label" or "Issuer:Label"
            const label = decodeURIComponent(url.pathname.replace(/^\/\/totp\//, ''));
            const name = label.includes(':') ? label.split(':')[1] : label;

            if (!secret) {
                Alert.alert('Invalid Code', 'No secret key found in QR code.', [
                    { text: 'OK', onPress: () => setScanned(false) }
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
                { text: 'OK', onPress: () => router.back() }
            ]);

        } catch (e) {
            console.error(e);
            Alert.alert('Error', 'Failed to parse QR code.', [
                { text: 'OK', onPress: () => setScanned(false) }
            ]);
        }
    };

    function toggleCameraFacing() {
        setFacing(current => (current === 'back' ? 'front' : 'back'));
    }

    return (
        <View style={styles.container}>
            <CameraView
                style={styles.camera}
                facing={facing}
                onBarcodeScanned={scanned ? undefined : handleBarCodeScanned}
                barcodeScannerSettings={{
                    barcodeTypes: ["qr"],
                }}
            >
                <View style={styles.buttonContainer}>
                    <TouchableOpacity style={styles.button} onPress={toggleCameraFacing}>
                        <Text style={styles.text}>Flip Camera</Text>
                    </TouchableOpacity>
                </View>
                <View style={styles.overlay}>
                    <View style={styles.scanFrame} />
                    <Text style={styles.overlayText}>Scan QR Code</Text>
                </View>
            </CameraView>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        justifyContent: 'center',
    },
    message: {
        textAlign: 'center',
        paddingBottom: 10,
    },
    camera: {
        flex: 1,
    },
    buttonContainer: {
        flex: 1,
        flexDirection: 'row',
        backgroundColor: 'transparent',
        margin: 64,
    },
    button: {
        flex: 1,
        alignSelf: 'flex-end',
        alignItems: 'center',
    },
    text: {
        fontSize: 24,
        fontWeight: 'bold',
        color: 'white',
    },
    overlay: {
        ...StyleSheet.absoluteFillObject,
        justifyContent: 'center',
        alignItems: 'center',
    },
    scanFrame: {
        width: 250,
        height: 250,
        borderWidth: 2,
        borderColor: '#007AFF',
        backgroundColor: 'transparent',
        marginBottom: 20,
    },
    overlayText: {
        color: 'white',
        fontSize: 18,
        backgroundColor: 'rgba(0,0,0,0.6)',
        padding: 8,
        borderRadius: 4,
    }
});
