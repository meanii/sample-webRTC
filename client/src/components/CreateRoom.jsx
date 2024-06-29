import React from 'react'
import { useNavigate } from 'react-router-dom'

export const CreateRoom = () => {
    const navigation = useNavigate()

    const create = async (e) => {
        e.preventDefault()
        const resp = await fetch('http://localhost:3000/create')
        const { room_id } = await resp.json()

       navigation(`/room/${room_id}`)
    }

    return (
        <div>
            <button onClick={create}>Create Room</button>
        </div> 
    )
}
