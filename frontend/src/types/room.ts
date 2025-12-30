export interface Participant {
  user_id: string;
  name: string;
  avatar: string;
  joined_at: string;
  left_at?: string;
}

export interface Room {
  id: string;
  created_by: string;
  status: 'active' | 'ended';
  participants: Participant[];
  max_capacity: number;
  created_at: string;
  ended_at?: string;
}

export interface JoinRoomRequest {
  user_name: string;
  avatar: string;
}
