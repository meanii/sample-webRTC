import React, { useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';

export const Room = () => {
    const userVideoRef = useRef();
    const userStreamRef = useRef();
    const partnerVideoRef = useRef();
    const peerRef = useRef();
    const webSocketRef = useRef();
    const { roomID } = useParams();

    const openCamera = async () => {
        try {
            const allDevices = await navigator.mediaDevices.enumerateDevices();
            const cameras = allDevices.filter(device => device.kind === 'videoinput');
            const constraints = {
                audio: true,
                video: {
                    deviceId: cameras?.[0]?.deviceId
                }
            };

            return await navigator.mediaDevices.getUserMedia(constraints);
        } catch (error) {
            console.log('App::Room::openCamera::Error:', error);
        }
    };

    useEffect(() => {
        const initWebSocket = async () => {
            const stream = await openCamera();
            if (stream) {
                userVideoRef.current.srcObject = stream;
                userStreamRef.current = stream;

                if (!webSocketRef.current) {
                    webSocketRef.current = new WebSocket(`ws://localhost:3000/join?roomID=${roomID}`);
                }

                webSocketRef.current.addEventListener('open', () => {
                    webSocketRef.current.send(JSON.stringify({ joined: true, roomID }));
                });

                webSocketRef.current.addEventListener('message', async (e) => {
                    const message = JSON.parse(e.data);
                    console.log('App::WebSocket::NewMessage:', e.data);

                    if (message?.joined) {
                        callUserForWebRTCExchange();
                    }

                    if (message?.offer) {
                        handlePeerConnOffer(message.offer);
                    }

                    if (message?.answer) {
                        console.log('App::Room::Event.message.answer::Info: Received Answer');
                        await peerRef.current.setRemoteDescription(
                            new RTCSessionDescription(message.answer)
                        );
                    }

                    if (message?.iceCandidate) {
                        console.log('App::Room::Event.message.iceCandidate::Info: Receiving and Adding ICE Candidate', message.iceCandidate);
                        try {
                            await peerRef.current.addIceCandidate(
                                message.iceCandidate
                            );
                        } catch (error) {
                            console.log('App::Room::Event.message.iceCandidate::Error:', error);
                        }
                    }
                });
            }
        };

        initWebSocket();

        return () => {
            if (webSocketRef.current) {
                webSocketRef.current.close();
            }
            if (peerRef.current) {
                peerRef.current.close();
            }
        };
    }, [roomID]);

    const handlePeerConnOffer = async (offer) => {
        console.log('App::Room::handlePeerConnOffer::Info: Received Offer and Creating Answer.', offer);
        peerRef.current = createWebRTCPeerConnection();

        await peerRef.current.setRemoteDescription(new RTCSessionDescription(offer));

        userStreamRef.current.getTracks().forEach(track => {
            peerRef.current.addTrack(track, userStreamRef.current);
        });

        const answer = await peerRef.current.createAnswer();
        await peerRef.current.setLocalDescription(answer);

        webSocketRef.current.send(
            JSON.stringify({ answer: peerRef.current.localDescription })
        );
    };

    const callUserForWebRTCExchange = () => {
        console.log('App::Room::callUserForWebRTCExchange::Info: Calling other users for connections.');
        peerRef.current = createWebRTCPeerConnection();

        userStreamRef.current.getTracks().forEach(track => {
            peerRef.current.addTrack(track, userStreamRef.current);
        });
    };

    const createWebRTCPeerConnection = () => {
        console.log('App::Room::createWebRTCPeerConnection::Info: Creating new WebRTC connection.');
        const peer = new RTCPeerConnection({
            iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
        });
        peer.onnegotiationneeded = handleNegotiationNeeded;
        peer.onicecandidate = handleIceCandidateEvent;
        peer.ontrack = handleTrackEvent;
        return peer;
    };

    const handleNegotiationNeeded = async () => {
        console.log('App::Room::handleNegotiationNeeded::Info: Creating offer');
        try {
            const myOffer = await peerRef.current.createOffer();
            await peerRef.current.setLocalDescription(myOffer);

            webSocketRef.current.send(
                JSON.stringify({ offer: peerRef.current.localDescription })
            );
        } catch (error) {
            console.log('App::Room::handleNegotiationNeeded::Error:', error);
        }
    };

    const handleIceCandidateEvent = (e) => {
        console.log('App::Room::handleIceCandidateEvent::Info: Found new ICE candidate');
        if (e.candidate) {
            console.log(e.candidate);
            webSocketRef.current.send(JSON.stringify({ iceCandidate: e.candidate }));
        }
    };

    const handleTrackEvent = (e) => {
        console.log('App::Room::handleTrackEvent::Info: Received track event');
        partnerVideoRef.current.srcObject = e.streams[0];
    };

    return (
        <div>
            <video autoPlay controls={true} ref={userVideoRef}></video>
            <video autoPlay controls={true} ref={partnerVideoRef}></video>
        </div>
    );
};
