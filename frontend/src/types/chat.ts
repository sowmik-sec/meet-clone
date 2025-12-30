export interface Message {
  id: string;
  room_id: string;
  user_id: string;
  user_name: string;
  message: string;
  timestamp: string;
}

export interface WSMessage {
  type: 'participant_joined' | 'participant_left' | 'chat_message' | 'room_ended';
  room_id: string;
  user_id: string;
  payload?: any;
}
